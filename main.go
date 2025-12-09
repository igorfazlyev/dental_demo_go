package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Data structures
type Scan struct {
	ID          int       `json:"id"`
	Date        string    `json:"date"`
	Status      string    `json:"status"`
	AIProcessed bool      `json:"ai_processed"`
}

type Procedure struct {
	Type     string `json:"type"`
	Position string `json:"position"`
	Urgency  string `json:"urgency"`
}

type TreatmentPlan struct {
	Diagnoses  []string    `json:"diagnoses"`
	Procedures []Procedure `json:"procedures"`
}

type ClinicOffer struct {
	Clinic      string  `json:"clinic"`
	Rating      float64 `json:"rating"`
	Cost        int     `json:"cost"`
	Duration    string  `json:"duration"`
	Warranty    string  `json:"warranty"`
	Installment string  `json:"installment"`
	Details     string  `json:"details"`
}

type IncomingPlan struct {
	ID         int    `json:"id"`
	Age        int    `json:"age"`
	Gender     string `json:"gender"`
	Date       string `json:"date"`
	Procedures string `json:"procedures"`
	Status     string `json:"status"`
}

type Lead struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Plan   string `json:"plan"`
	Cost   int    `json:"cost"`
	Status string `json:"status"`
}

type Session struct {
	UserRole       string
	LoggedIn       bool
	SelectedClinic string
	PatientScans   []Scan
	TreatmentPlan  TreatmentPlan
	ClinicOffers   []ClinicOffer
	IncomingPlans  []IncomingPlan
	Leads          []Lead
}

// Global session storage
var (
	sessions  = make(map[string]*Session)
	mu        sync.RWMutex
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Initialize mock data
func initSession() *Session {
	return &Session{
		LoggedIn: false,
		PatientScans: []Scan{
			{ID: 1, Date: "2025-11-15", Status: "ready", AIProcessed: true},
			{ID: 2, Date: "2025-12-05", Status: "processing", AIProcessed: false},
		},
		TreatmentPlan: TreatmentPlan{
			Diagnoses: []string{
				"Кариес зуба 1.6",
				"Пульпит зуба 2.5",
				"Отсутствует зуб 3.7",
			},
			Procedures: []Procedure{
				{Type: "Имплант", Position: "3.7", Urgency: "Средняя"},
				{Type: "Коронка", Position: "2.5", Urgency: "Высокая"},
				{Type: "Пломбирование", Position: "1.6", Urgency: "Высокая"},
			},
		},
		ClinicOffers: []ClinicOffer{
			{
				Clinic:      "СтомаПрофи",
				Rating:      4.8,
				Cost:        185000,
				Duration:    "3-4 месяца",
				Warranty:    "5 лет на имплант",
				Installment: "До 12 месяцев",
				Details:     "Имплант Nobel - 95000₽, Коронка - 35000₽, Лечение каналов - 15000₽, Прочее - 40000₽",
			},
			{
				Clinic:      "Дентал Плюс",
				Rating:      4.5,
				Cost:        165000,
				Duration:    "2-3 месяца",
				Warranty:    "3 года на имплант",
				Installment: "До 6 месяцев",
				Details:     "Имплант Osstem - 75000₽, Коронка - 30000₽, Лечение каналов - 12000₽, Прочее - 48000₽",
			},
			{
				Clinic:      "ЭлитДент",
				Rating:      4.9,
				Cost:        225000,
				Duration:    "3-5 месяцев",
				Warranty:    "10 лет на имплант",
				Installment: "До 24 месяцев",
				Details:     "Имплант Straumann - 120000₽, Коронка - 45000₽, Лечение каналов - 20000₽, Прочее - 40000₽",
			},
		},
		IncomingPlans: []IncomingPlan{
			{ID: 1, Age: 35, Gender: "Ж", Date: "2025-12-08", Procedures: "Имплант 3.7, Коронка 2.5, Пломбирование 1.6", Status: "new"},
			{ID: 2, Age: 42, Gender: "М", Date: "2025-12-07", Procedures: "Протезирование верхняя челюсть", Status: "offer_sent"},
			{ID: 3, Age: 28, Gender: "Ж", Date: "2025-12-09", Procedures: "Кариес множественный, 4 пломбы", Status: "new"},
		},
		Leads: []Lead{
			{ID: 1, Name: "Анна Петрова", Phone: "+7 916 555-1234", Plan: "Имплант 3.7, Коронка 2.5", Cost: 165000, Status: "Не обработан"},
			{ID: 2, Name: "Игорь Смирнов", Phone: "+7 926 555-5678", Plan: "Протезирование", Cost: 280000, Status: "Записан на консультацию"},
		},
	}
}

func getSession(r *http.Request) *Session {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return initSession()
	}

	mu.RLock()
	defer mu.RUnlock()

	if sess, ok := sessions[cookie.Value]; ok {
		return sess
	}
	return initSession()
}

func saveSession(w http.ResponseWriter, r *http.Request, sess *Session) {
	cookie, err := r.Cookie("session_id")
	sessionID := ""

	if err != nil {
		sessionID = fmt.Sprintf("%d", rand.Int63())
		http.SetCookie(w, &http.Cookie{
			Name:   "session_id",
			Value:  sessionID,
			Path:   "/",
			MaxAge: 86400,
		})
	} else {
		sessionID = cookie.Value
	}

	mu.Lock()
	sessions[sessionID] = sess
	mu.Unlock()
}

// Template rendering helper with proper layout support
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// Parse the base layout and the specific template
	t, err := template.ParseFiles("templates/base.html", "templates/"+tmpl)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the base layout (which will call the "content" template)
	err = t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmpl, err)
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)

	if !sess.LoggedIn {
		renderTemplate(w, "login.html", nil)
		return
	}

	switch sess.UserRole {
	case "patient":
		http.Redirect(w, r, "/patient/scans", http.StatusSeeOther)
	case "clinic":
		http.Redirect(w, r, "/clinic/dashboard", http.StatusSeeOther)
	case "government":
		http.Redirect(w, r, "/government/dashboard", http.StatusSeeOther)
	default:
		renderTemplate(w, "login.html", nil)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		role := r.FormValue("role")
		sess := getSession(r)
		sess.LoggedIn = true
		sess.UserRole = role
		saveSession(w, r, sess)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	sess.LoggedIn = false
	sess.UserRole = ""
	saveSession(w, r, sess)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Patient handlers
func patientScansHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "patient_scans.html", sess)
}

func patientPlanHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "patient_plan.html", sess)
}

func patientCriteriaHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "patient_criteria.html", sess)
}

func patientOffersHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		clinic := r.FormValue("clinic")
		sess.SelectedClinic = clinic
		saveSession(w, r, sess)
		http.Redirect(w, r, "/patient/consultations", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "patient_offers.html", sess)
}

func patientConsultationsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "patient_consultations.html", sess)
}

func patientReviewsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "patient" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "patient_reviews.html", sess)
}

// Clinic handlers
func clinicDashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "clinic_dashboard.html", sess)
}

func clinicPlansHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		action := r.FormValue("action")
		planID, _ := strconv.Atoi(r.FormValue("plan_id"))

		for i, plan := range sess.IncomingPlans {
			if plan.ID == planID {
				if action == "calculate" {
					sess.IncomingPlans[i].Status = "calculated"
				} else if action == "send" {
					sess.IncomingPlans[i].Status = "offer_sent"
				}
				break
			}
		}
		saveSession(w, r, sess)
		http.Redirect(w, r, "/clinic/plans", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "clinic_plans.html", sess)
}

func clinicLeadsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "clinic_leads.html", sess)
}

func clinicAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "clinic_analytics.html", sess)
}

func clinicPricelistHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "clinic_pricelist.html", sess)
}

// Government handlers
func governmentDashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "government" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "government_dashboard.html", sess)
}

func governmentAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "government" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "government_analytics.html", sess)
}

// API handlers
func apiCalculateHandler(w http.ResponseWriter, r *http.Request) {
	cost := 150000 + rand.Intn(100000)
	response := map[string]interface{}{
		"success": true,
		"cost":    cost,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	// Patient routes
	http.HandleFunc("/patient/scans", patientScansHandler)
	http.HandleFunc("/patient/plan", patientPlanHandler)
	http.HandleFunc("/patient/criteria", patientCriteriaHandler)
	http.HandleFunc("/patient/offers", patientOffersHandler)
	http.HandleFunc("/patient/consultations", patientConsultationsHandler)
	http.HandleFunc("/patient/reviews", patientReviewsHandler)

	// Clinic routes
	http.HandleFunc("/clinic/dashboard", clinicDashboardHandler)
	http.HandleFunc("/clinic/plans", clinicPlansHandler)
	http.HandleFunc("/clinic/leads", clinicLeadsHandler)
	http.HandleFunc("/clinic/analytics", clinicAnalyticsHandler)
	http.HandleFunc("/clinic/pricelist", clinicPricelistHandler)

	// Government routes
	http.HandleFunc("/government/dashboard", governmentDashboardHandler)
	http.HandleFunc("/government/analytics", governmentAnalyticsHandler)

	// API routes
	http.HandleFunc("/api/calculate", apiCalculateHandler)

	fmt.Println("🦷 DentalAI Platform starting...")
	fmt.Println("🌐 Server running on http://localhost:8080")
	fmt.Println("📱 Open your browser and navigate to http://localhost:8080")
	fmt.Println("")
	fmt.Println("✨ Using layout-based templates with base.html")
	fmt.Println("🎨 Static files served from static/ directory")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
