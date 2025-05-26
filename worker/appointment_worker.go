package worker

import (
	"biostat/service"
	"log"
	"time"
)

func StartAppointmentScheduler(service service.AppointmentService) {
	log.Println("Appointment Schedular running")

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			processAppointments(service)
		}
	}()
}

func processAppointments(s service.AppointmentService) {
	log.Println("Scheduler: checking for completed appointments ...")
	err := s.MarkCompletedAppointments()
	if err != nil {
		log.Println("Error @ MarkCompletedAppointments", err)
	}
}
