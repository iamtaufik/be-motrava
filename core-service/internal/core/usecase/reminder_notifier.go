package usecase

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"motrava/core-service/internal/core/models"
	fcmPort "motrava/core-service/internal/core/port/fcm"
	"motrava/core-service/internal/core/repository"
)

type ReminderNotifier struct {
	fcmClient  fcmPort.FCMClient
	deviceRepo repository.UserDeviceRepository
}

func NewReminderNotifier(fcmClient fcmPort.FCMClient, deviceRepo repository.UserDeviceRepository) *ReminderNotifier {
	return &ReminderNotifier{fcmClient: fcmClient, deviceRepo: deviceRepo}
}

func (n *ReminderNotifier) CheckAndNotify(reminder *models.ServiceReminder) {
	if reminder.IntervalKM <= 0 {
		return
	}

	progressPercent := math.Round((reminder.AccumulatedKM/reminder.IntervalKM)*100*100) / 100

	if progressPercent < ServiceReminderThresholdPercent {
		return
	}

	if reminder.NotifiedAt != nil {
		if reminder.LastServiceAt != nil && reminder.NotifiedAt.After(*reminder.LastServiceAt) {
			return
		}
	}

	devices, err := n.deviceRepo.FindByUserID(reminder.UserID)
	if err != nil || len(devices) == 0 {
		return
	}

	tokens := make([]string, len(devices))
	for i, d := range devices {
		tokens[i] = d.DeviceToken
	}

	title := "Service Reminder"
	body := fmt.Sprintf("%s sudah %.0f%% — %.0f/%.0f KM. Segera service kendaraan Anda.",
		reminder.ServiceName, progressPercent, reminder.AccumulatedKM, reminder.IntervalKM)

	data := map[string]string{
		"type":             "service_reminder",
		"reminder_id":      reminder.ID.String(),
		"vehicle_id":       reminder.VehicleID.String(),
		"service_name":     reminder.ServiceName,
		"progress_percent": fmt.Sprintf("%.0f", progressPercent),
		"accumulated_km":   fmt.Sprintf("%.0f", reminder.AccumulatedKM),
		"interval_km":      fmt.Sprintf("%.0f", reminder.IntervalKM),
	}

	notification := fcmPort.FCMNotification{
		Title: title,
		Body:  body,
		Data:  data,
	}

	n.fcmClient.SendMulticast(tokens, notification)

	now := time.Now().UTC()
	reminder.NotifiedAt = &now
}

func (n *ReminderNotifier) NotifyAllActive(reminders []models.ServiceReminder) {
	for i := range reminders {
		n.CheckAndNotify(&reminders[i])
	}
}

func (n *ReminderNotifier) ProcessAllActive(reminderRepo repository.ServiceReminderRepository) {
	reminders, err := reminderRepo.FindAllActive()
	if err != nil {
		return
	}

	now := time.Now().UTC()
	for i := range reminders {
		r := &reminders[i]

		if r.IntervalKM <= 0 {
			continue
		}

		progressPercent := math.Round((r.AccumulatedKM/r.IntervalKM)*100*100) / 100
		if progressPercent < ServiceReminderThresholdPercent {
			continue
		}

		if r.NotifiedAt != nil {
			if r.LastServiceAt != nil && r.NotifiedAt.After(*r.LastServiceAt) {
				continue
			}
		}

		devices, err := n.deviceRepo.FindByUserID(r.UserID)
		if err != nil || len(devices) == 0 {
			continue
		}

		tokens := make([]string, len(devices))
		for j, d := range devices {
			tokens[j] = d.DeviceToken
		}

		title := "Service Reminder"
		body := fmt.Sprintf("%s sudah %.0f%% — %.0f/%.0f KM. Segera service kendaraan Anda.",
			r.ServiceName, progressPercent, r.AccumulatedKM, r.IntervalKM)

		data := map[string]string{
			"type":             "service_reminder",
			"reminder_id":      r.ID.String(),
			"vehicle_id":       r.VehicleID.String(),
			"service_name":     r.ServiceName,
			"progress_percent": fmt.Sprintf("%.0f", progressPercent),
			"accumulated_km":   fmt.Sprintf("%.0f", r.AccumulatedKM),
			"interval_km":      fmt.Sprintf("%.0f", r.IntervalKM),
		}

		notification := fcmPort.FCMNotification{
			Title: title,
			Body:  body,
			Data:  data,
		}

		n.fcmClient.SendMulticast(tokens, notification)

		r.NotifiedAt = &now
		reminderRepo.Save(r)
	}
}

// ReminderScheduler periodically processes active reminders. It prevents
// overlapping runs and can be stopped gracefully.
type ReminderScheduler struct {
	notifier     *ReminderNotifier
	reminderRepo repository.ServiceReminderRepository
	stopCh       chan struct{}
	done         chan struct{}
	running      atomic.Bool
}

func StartReminderScheduler(interval time.Duration, notifier *ReminderNotifier, reminderRepo repository.ServiceReminderRepository) *ReminderScheduler {
	s := &ReminderScheduler{
		notifier:     notifier,
		reminderRepo: reminderRepo,
		stopCh:       make(chan struct{}),
		done:         make(chan struct{}),
	}

	go func() {
		defer close(s.done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.runOnce()
			}
		}
	}()

	return s
}

func (s *ReminderScheduler) runOnce() {
	// Skip this tick if the previous run is still in progress so runs never
	// overlap and send duplicate notifications.
	if !s.running.CompareAndSwap(false, true) {
		return
	}
	defer s.running.Store(false)

	s.notifier.ProcessAllActive(s.reminderRepo)
}

// Stop halts the scheduler and waits for any in-flight run to finish.
func (s *ReminderScheduler) Stop() {
	close(s.stopCh)
	<-s.done
}

func UpdateNotifiedAt(reminder *models.ServiceReminder, notifier *ReminderNotifier, reminderRepo repository.ServiceReminderRepository) {
	if reminder.IntervalKM <= 0 {
		return
	}

	progressPercent := math.Round((reminder.AccumulatedKM/reminder.IntervalKM)*100*100) / 100
	if progressPercent < ServiceReminderThresholdPercent {
		return
	}

	if reminder.NotifiedAt != nil {
		if reminder.LastServiceAt != nil && reminder.NotifiedAt.After(*reminder.LastServiceAt) {
			return
		}
	}

	now := time.Now().UTC()
	reminder.NotifiedAt = &now
	reminderRepo.Save(reminder)

	notifier.CheckAndNotify(reminder)
}

func AddTripDistanceToReminders(tripDistanceMeters float64, vehicleID uuid.UUID, reminderRepo repository.ServiceReminderRepository, notifier *ReminderNotifier) {
	if tripDistanceMeters <= 0 {
		return
	}

	tripKM := math.Round((tripDistanceMeters/1000.0)*100) / 100

	reminders, err := reminderRepo.FindByVehicleID(vehicleID)
	if err != nil {
		return
	}

	for i := range reminders {
		r := &reminders[i]
		if !r.IsActive {
			continue
		}

		r.AccumulatedKM = math.Round((r.AccumulatedKM+tripKM)*100) / 100

		if notifier != nil {
			notifier.CheckAndNotify(r)
		}

		reminderRepo.Save(r)
	}
}
