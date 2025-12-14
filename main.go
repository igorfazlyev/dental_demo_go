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
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	AIProcessed bool   `json:"ai_processed"`
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
	Notes  string `json:"notes"` // ADD THIS LINE
}

type PriceItem struct {
	Service  string `json:"service"`
	Price    int    `json:"price"`
	Warranty int    `json:"warranty"`
}

// type Session struct {
// 	UserRole       string
// 	LoggedIn       bool
// 	LoginError     string
// 	SelectedClinic string
// 	PatientScans   []Scan
// 	TreatmentPlan  TreatmentPlan
// 	ClinicOffers   []ClinicOffer
// 	IncomingPlans  []IncomingPlan
// 	Leads          []Lead
// }

type Session struct {
	UserRole         string
	LoggedIn         bool
	LoginError       string
	SelectedClinic   string
	PatientScans     []Scan
	TreatmentPlan    TreatmentPlan
	ClinicOffers     []ClinicOffer
	IncomingPlans    []IncomingPlan
	Leads            []Lead
	PriceList        []PriceItem
	SuccessMessage   string
	AnalyticsFilters AnalyticsFilters // ADD THIS
	AnalyticsData    AnalyticsData    // ADD THIS
	DiseaseStats     []DiseaseStats   // ADD THIS
}

type AnalyticsData struct {
	Clinic         string
	PlansProcessed int
	AverageCheck   int
	ConversionRate int
	MonthlyData    []MonthData
}

type MonthData struct {
	Month    string
	Revenue  float64
	Patients int
}

type DiseaseStats struct {
	Disease    string
	Cases      int
	AverageAge int
}

type AnalyticsFilters struct {
	Clinic   string
	AgeGroup string
	Gender   string
}

// Demo user credentials
type User struct {
	Username string
	Password string
	Role     string
}

var demoUsers = []User{
	{Username: "patient", Password: "demo123", Role: "patient"},
	{Username: "clinic", Password: "demo123", Role: "clinic"},
	{Username: "government", Password: "demo123", Role: "government"},
}

// Global session storage
var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Initialize mock data with realistic 2-3 months of activity
func initSession() *Session {
	return &Session{
		LoggedIn: false,

		// Patient: 10 scans over 2.5 months
		PatientScans: []Scan{
			{ID: 1, Date: "2025-09-15", Status: "ready", AIProcessed: true},
			{ID: 2, Date: "2025-09-28", Status: "ready", AIProcessed: true},
			{ID: 3, Date: "2025-10-12", Status: "ready", AIProcessed: true},
			{ID: 4, Date: "2025-10-25", Status: "ready", AIProcessed: true},
			{ID: 5, Date: "2025-11-08", Status: "ready", AIProcessed: true},
			{ID: 6, Date: "2025-11-15", Status: "ready", AIProcessed: true},
			{ID: 7, Date: "2025-11-28", Status: "ready", AIProcessed: true},
			{ID: 8, Date: "2025-12-05", Status: "processing", AIProcessed: false},
			{ID: 9, Date: "2025-12-08", Status: "processing", AIProcessed: false},
			{ID: 10, Date: "2025-12-09", Status: "processing", AIProcessed: false},
		},

		// Patient: 15 diagnoses and 10 procedures
		TreatmentPlan: TreatmentPlan{
			Diagnoses: []string{
				"Кариес зуба 1.6 (глубокий)",
				"Пульпит зуба 2.5 (острый)",
				"Периодонтит зуба 3.7 (хронический)",
				"Отсутствует зуб 3.7",
				"Кариес зуба 4.6 (средний)",
				"Гингивит (генерализованный)",
				"Пародонтит средней степени",
				"Кариес зуба 1.4 (поверхностный)",
				"Скол коронки зуба 2.1",
				"Отсутствует зуб 4.5",
				"Кариес корня зуба 3.6",
				"Подвижность зуба 2.8 (II степень)",
				"Дефект пломбы зуба 1.7",
				"Клиновидный дефект зуба 1.3",
				"Зубной камень (множественный)",
			},
			Procedures: []Procedure{
				{Type: "Имплант", Position: "3.7", Urgency: "Средняя"},
				{Type: "Имплант", Position: "4.5", Urgency: "Низкая"},
				{Type: "Коронка", Position: "2.5", Urgency: "Высокая"},
				{Type: "Коронка", Position: "2.1", Urgency: "Средняя"},
				{Type: "Лечение каналов", Position: "2.5", Urgency: "Высокая"},
				{Type: "Лечение каналов", Position: "3.7", Urgency: "Высокая"},
				{Type: "Пломбирование", Position: "1.6", Urgency: "Высокая"},
				{Type: "Пломбирование", Position: "4.6", Urgency: "Средняя"},
				{Type: "Пломбирование", Position: "1.4", Urgency: "Низкая"},
				{Type: "Профессиональная чистка", Position: "Все зубы", Urgency: "Средняя"},
			},
		},

		// Patient: 3 clinic offers
		ClinicOffers: []ClinicOffer{
			{
				Clinic:      "СтомаПрофи",
				Rating:      4.8,
				Cost:        185000,
				Duration:    "3-4 месяца",
				Warranty:    "5 лет на имплант",
				Installment: "До 12 месяцев",
				Details:     "Имплант Nobel - 95000₽, Коронка - 35000₽, Лечение каналов - 15000₽, Пломбы (3шт) - 15000₽, Чистка - 5000₽, Прочее - 20000₽",
			},
			{
				Clinic:      "Дентал Плюс",
				Rating:      4.5,
				Cost:        165000,
				Duration:    "2-3 месяца",
				Warranty:    "3 года на имплант",
				Installment: "До 6 месяцев",
				Details:     "Имплант Osstem - 75000₽, Коронка - 30000₽, Лечение каналов - 12000₽, Пломбы (3шт) - 12000₽, Чистка - 4000₽, Прочее - 32000₽",
			},
			{
				Clinic:      "ЭлитДент",
				Rating:      4.9,
				Cost:        225000,
				Duration:    "3-5 месяцев",
				Warranty:    "10 лет на имплант",
				Installment: "До 24 месяцев",
				Details:     "Имплант Straumann - 120000₽, Коронка - 45000₽, Лечение каналов - 20000₽, Пломбы (3шт) - 18000₽, Чистка - 7000₽, Прочее - 15000₽",
			},
		},

		// Clinic: 15 incoming plans over 2 months
		IncomingPlans: []IncomingPlan{
			{ID: 1, Age: 35, Gender: "Ж", Date: "2025-12-09", Procedures: "Имплант 3.7, Коронка 2.5, Лечение каналов", Status: "new"},
			{ID: 2, Age: 42, Gender: "М", Date: "2025-12-08", Procedures: "Протезирование верхняя челюсть (6 единиц)", Status: "offer_sent"},
			{ID: 3, Age: 28, Gender: "Ж", Date: "2025-12-08", Procedures: "Кариес множественный, 4 пломбы", Status: "new"},
			{ID: 4, Age: 51, Gender: "М", Date: "2025-12-07", Procedures: "Имплант 4.6, 4.7, костная пластика", Status: "calculated"},
			{ID: 5, Age: 33, Gender: "Ж", Date: "2025-12-06", Procedures: "Лечение каналов 1.6, коронка", Status: "offer_sent"},
			{ID: 6, Age: 46, Gender: "М", Date: "2025-12-05", Procedures: "Пародонтологическое лечение комплексное", Status: "offer_sent"},
			{ID: 7, Age: 29, Gender: "Ж", Date: "2025-12-04", Procedures: "Эстетическая реставрация 4 передних зуба", Status: "calculated"},
			{ID: 8, Age: 38, Gender: "М", Date: "2025-12-03", Procedures: "Удаление зуба мудрости + имплант 3.6", Status: "new"},
			{ID: 9, Age: 44, Gender: "Ж", Date: "2025-12-02", Procedures: "Виниры 6 единиц", Status: "offer_sent"},
			{ID: 10, Age: 55, Gender: "М", Date: "2025-11-30", Procedures: "Полное протезирование нижняя челюсть", Status: "expired"},
			{ID: 11, Age: 31, Gender: "Ж", Date: "2025-11-28", Procedures: "Лечение каналов 2.5, 2.6", Status: "offer_sent"},
			{ID: 12, Age: 48, Gender: "М", Date: "2025-11-25", Procedures: "Имплант 1.6, синус-лифтинг", Status: "offer_sent"},
			{ID: 13, Age: 26, Gender: "Ж", Date: "2025-11-22", Procedures: "Отбеливание + профчистка", Status: "expired"},
			{ID: 14, Age: 53, Gender: "М", Date: "2025-11-18", Procedures: "Имплант 3.7, 4.7, протезирование мостовидное", Status: "offer_sent"},
			{ID: 15, Age: 37, Gender: "Ж", Date: "2025-11-15", Procedures: "Лечение пульпита 3шт, пломбирование 2шт", Status: "offer_sent"},
		},

		// Clinic: 10 leads with Russian names
		Leads: []Lead{
			{ID: 1, Name: "Анна Петрова", Phone: "+7 916 555-1234", Plan: "Имплант 3.7, Коронка 2.5", Cost: 165000, Status: "Не обработан", Notes: ""},
			{ID: 2, Name: "Игорь Смирнов", Phone: "+7 926 555-5678", Plan: "Протезирование верхняя челюсть", Cost: 280000, Status: "Записан на консультацию", Notes: ""},
			{ID: 3, Name: "Елена Ковалева", Phone: "+7 905 555-2341", Plan: "Имплант 4.6, 4.7 + костная пластика", Cost: 320000, Status: "Записан на консультацию", Notes: ""},
			{ID: 4, Name: "Дмитрий Волков", Phone: "+7 903 555-8765", Plan: "Лечение каналов + коронка", Cost: 45000, Status: "Лечение начато", Notes: ""},
			{ID: 5, Name: "Мария Соколова", Phone: "+7 915 555-4567", Plan: "Виниры 6 единиц", Cost: 180000, Status: "Не обработан", Notes: ""},
			{ID: 6, Name: "Сергей Морозов", Phone: "+7 925 555-7890", Plan: "Пародонтологическое лечение", Cost: 85000, Status: "Лечение начато", Notes: ""},
			{ID: 7, Name: "Ольга Новикова", Phone: "+7 917 555-3456", Plan: "Эстетическая реставрация 4 зуба", Cost: 96000, Status: "Записан на консультацию", Notes: ""},
			{ID: 8, Name: "Александр Лебедев", Phone: "+7 906 555-6543", Plan: "Имплант + синус-лифтинг", Cost: 195000, Status: "Не обработан", Notes: ""},
			{ID: 9, Name: "Татьяна Козлова", Phone: "+7 916 555-9876", Plan: "Лечение каналов 2.5, 2.6", Cost: 28000, Status: "Лечение завершено", Notes: ""},
			{ID: 10, Name: "Владимир Орлов", Phone: "+7 929 555-1111", Plan: "Полное протезирование нижняя челюсть", Cost: 450000, Status: "Отказ", Notes: "Слишком дорого"},
		},
		PriceList: []PriceItem{
			{Service: "Имплант (Nobel)", Price: 95000, Warranty: 5},
			{Service: "Имплант (Osstem)", Price: 75000, Warranty: 3},
			{Service: "Имплант (Straumann)", Price: 120000, Warranty: 10},
			{Service: "Коронка металлокерамика", Price: 30000, Warranty: 2},
			{Service: "Коронка циркониевая", Price: 45000, Warranty: 5},
			{Service: "Лечение каналов", Price: 15000, Warranty: 1},
			{Service: "Пломба светоотверждаемая", Price: 5000, Warranty: 1},
			{Service: "Удаление зуба", Price: 3000, Warranty: 0},
			{Service: "Профессиональная чистка", Price: 4000, Warranty: 0},
			{Service: "Отбеливание", Price: 15000, Warranty: 0},
		},
	}
}

// Validate user credentials
func validateCredentials(username, password string) (string, bool) {
	for _, user := range demoUsers {
		if user.Username == username && user.Password == password {
			return user.Role, true
		}
	}
	return "", false
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
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, err := template.ParseFiles("templates/base.html", "templates/"+tmpl)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmpl, err)
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}
//

// REPLACE your getAnalyticsData() function with this enhanced version

func getAnalyticsData(clinic, ageGroup, gender string) (AnalyticsData, []DiseaseStats) {
    data := AnalyticsData{
        Clinic: clinic,
    }

    // ===== BASE VALUES =====
    basePlans := 89
    baseCheck := 175000
    baseConversion := 38
    baseRevenue := 22.5
    basePatients := 285

    // ===== CLINIC ADJUSTMENTS =====
    clinicMultiplier := 1.0
    if clinic == "ЭлитДент" {
        basePlans = 124
        baseCheck = 225000
        baseConversion = 45
        clinicMultiplier = 1.35
    } else if clinic == "Дентал Плюс" {
        basePlans = 156
        baseCheck = 165000
        baseConversion = 42
        clinicMultiplier = 1.15
    } else if clinic == "Все клиники" {
        basePlans = 369
        baseCheck = 188000
        baseConversion = 41
        clinicMultiplier = 3.2
    }

    // ===== AGE GROUP ADJUSTMENTS =====
    ageMultiplier := 1.0
    if ageGroup == "46-60" {
        baseCheck = int(float64(baseCheck) * 1.3)
        ageMultiplier = 1.25
    } else if ageGroup == "18-30" {
        baseCheck = int(float64(baseCheck) * 0.8)
        ageMultiplier = 0.85
    } else if ageGroup == "31-45" {
        baseCheck = int(float64(baseCheck) * 1.1)
        ageMultiplier = 1.1
    } else if ageGroup == "60+" {
        baseCheck = int(float64(baseCheck) * 1.4)
        ageMultiplier = 1.15
    }

    // ===== GENDER ADJUSTMENTS =====
    genderMultiplier := 1.0
    if gender == "Женский" {
        basePlans = int(float64(basePlans) * 1.2)
        genderMultiplier = 1.2
    } else if gender == "Мужской" {
        basePlans = int(float64(basePlans) * 0.9)
        genderMultiplier = 0.9
    }

    data.PlansProcessed = basePlans
    data.AverageCheck = baseCheck
    data.ConversionRate = baseConversion

    // ===== DYNAMIC MONTHLY DATA =====
    totalMultiplier := clinicMultiplier * ageMultiplier * genderMultiplier

    data.MonthlyData = []MonthData{
        {
            Month:    "Август",
            Revenue:  baseRevenue * totalMultiplier * 0.95,
            Patients: int(float64(basePatients) * totalMultiplier * 0.92),
        },
        {
            Month:    "Сентябрь",
            Revenue:  baseRevenue * totalMultiplier * 1.02,
            Patients: int(float64(basePatients) * totalMultiplier * 0.98),
        },
        {
            Month:    "Октябрь",
            Revenue:  baseRevenue * totalMultiplier * 1.0,
            Patients: int(float64(basePatients) * totalMultiplier * 0.95),
        },
        {
            Month:    "Ноябрь",
            Revenue:  baseRevenue * totalMultiplier * 1.08,
            Patients: int(float64(basePatients) * totalMultiplier * 1.05),
        },
        {
            Month:    "Декабрь",
            Revenue:  baseRevenue * totalMultiplier * 1.15,
            Patients: int(float64(basePatients) * totalMultiplier * 1.12),
        },
    }

    // ===== DYNAMIC DISEASE STATISTICS =====
    diseases := []DiseaseStats{
        {Disease: "Кариес", Cases: 523, AverageAge: 34},
        {Disease: "Пульпит", Cases: 224, AverageAge: 38},
        {Disease: "Пародонтит", Cases: 187, AverageAge: 45},
        {Disease: "Периодонтит", Cases: 98, AverageAge: 42},
        {Disease: "Гингивит", Cases: 156, AverageAge: 29},
    }

    // Adjust based on clinic size
    if clinic == "ЭлитДент" {
        for i := range diseases {
            diseases[i].Cases = int(float64(diseases[i].Cases) * 1.3)
        }
    } else if clinic == "Дентал Плюс" {
        for i := range diseases {
            diseases[i].Cases = int(float64(diseases[i].Cases) * 1.5)
        }
    } else if clinic == "Все клиники" {
        for i := range diseases {
            diseases[i].Cases = int(float64(diseases[i].Cases) * 3.5)
        }
    }

    // Adjust based on age group
    if ageGroup == "46-60" {
        diseases[2].Cases = int(float64(diseases[2].Cases) * 2.2) // Parodontitis
        diseases[3].Cases = int(float64(diseases[3].Cases) * 1.8) // Periodontitis
        diseases[0].Cases = int(float64(diseases[0].Cases) * 0.7) // Less caries
        diseases[4].Cases = int(float64(diseases[4].Cases) * 0.5) // Less gingivitis
        for i := range diseases {
            diseases[i].AverageAge = 52
        }
    } else if ageGroup == "18-30" {
        diseases[0].Cases = int(float64(diseases[0].Cases) * 1.8) // More caries
        diseases[4].Cases = int(float64(diseases[4].Cases) * 1.6) // More gingivitis
        diseases[2].Cases = int(float64(diseases[2].Cases) * 0.4) // Less parodontitis
        diseases[3].Cases = int(float64(diseases[3].Cases) * 0.3) // Less periodontitis
        for i := range diseases {
            diseases[i].AverageAge = 25
        }
    } else if ageGroup == "31-45" {
        diseases[0].Cases = int(float64(diseases[0].Cases) * 1.2)
        diseases[1].Cases = int(float64(diseases[1].Cases) * 1.3)
        for i := range diseases {
            diseases[i].AverageAge = 38
        }
    } else if ageGroup == "60+" {
        diseases[2].Cases = int(float64(diseases[2].Cases) * 2.8)
        diseases[3].Cases = int(float64(diseases[3].Cases) * 2.2)
        diseases[0].Cases = int(float64(diseases[0].Cases) * 0.5)
        for i := range diseases {
            diseases[i].AverageAge = 68
        }
    }

    // Adjust based on gender
    if gender == "Женский" {
        for i := range diseases {
            diseases[i].Cases = int(float64(diseases[i].Cases) * 1.15)
        }
    } else if gender == "Мужской" {
        for i := range diseases {
            diseases[i].Cases = int(float64(diseases[i].Cases) * 0.85)
        }
    }

    return data, diseases
}
// Handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)

	if !sess.LoggedIn {
		renderTemplate(w, "login.html", sess)
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
		renderTemplate(w, "login.html", sess)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate credentials
		role, valid := validateCredentials(username, password)

		if valid {
			sess := getSession(r)
			sess.LoggedIn = true
			sess.UserRole = role
			sess.LoginError = ""
			saveSession(w, r, sess)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Invalid credentials - show error
		sess := getSession(r)
		sess.LoginError = "Неверный логин или пароль"
		sess.LoggedIn = false
		saveSession(w, r, sess)
		renderTemplate(w, "login.html", sess)
		return
	}

	// GET request - show login form
	sess := getSession(r)
	sess.LoginError = ""
	renderTemplate(w, "login.html", sess)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	sess.LoggedIn = false
	sess.UserRole = ""
	sess.LoginError = ""
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

//	func clinicLeadsHandler(w http.ResponseWriter, r *http.Request) {
//		sess := getSession(r)
//		if !sess.LoggedIn || sess.UserRole != "clinic" {
//			http.Redirect(w, r, "/", http.StatusSeeOther)
//			return
//		}
//		renderTemplate(w, "clinic_leads.html", sess)
//	}

func clinicLeadsHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Handle POST request to update lead status/notes
	if r.Method == "POST" {
		leadID, _ := strconv.Atoi(r.FormValue("lead_id"))
		newStatus := r.FormValue("status")
		newNotes := r.FormValue("notes")

		// Find and update the lead
		for i, lead := range sess.Leads {
			if lead.ID == leadID {
				sess.Leads[i].Status = newStatus
				sess.Leads[i].Notes = newNotes
				break
			}
		}

		// Save the session
		saveSession(w, r, sess)

		// Redirect back to the page
		http.Redirect(w, r, "/clinic/leads", http.StatusSeeOther)
		return
	}

	// GET request - show the page
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

// func clinicPricelistHandler(w http.ResponseWriter, r *http.Request) {
// 	sess := getSession(r)
// 	if !sess.LoggedIn || sess.UserRole != "clinic" {
// 		http.Redirect(w, r, "/", http.StatusSeeOther)
// 		return
// 	}
// 	renderTemplate(w, "clinic_pricelist.html", sess)
// }

func clinicPricelistHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sess.SuccessMessage = ""

	if r.Method == "POST" {
		r.ParseForm()

		count, _ := strconv.Atoi(r.FormValue("count"))

		newPriceList := make([]PriceItem, 0)
		for i := 0; i < count; i++ {
			service := r.FormValue(fmt.Sprintf("service_%d", i))
			price, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("price_%d", i)))
			warranty, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("warranty_%d", i)))

			if service != "" {
				newPriceList = append(newPriceList, PriceItem{
					Service:  service,
					Price:    price,
					Warranty: warranty,
				})
			}
		}

		sess.PriceList = newPriceList
		sess.SuccessMessage = "Изменения успешно сохранены!"
		saveSession(w, r, sess)

		http.Redirect(w, r, "/clinic/pricelist", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "clinic_pricelist.html", sess)
}

func clinicPricelistAddHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		service := r.FormValue("service")
		price, _ := strconv.Atoi(r.FormValue("price"))
		warranty, _ := strconv.Atoi(r.FormValue("warranty"))

		newItem := PriceItem{
			Service:  service,
			Price:    price,
			Warranty: warranty,
		}

		sess.PriceList = append(sess.PriceList, newItem)
		sess.SuccessMessage = fmt.Sprintf("Услуга \"%s\" успешно добавлена!", service)
		saveSession(w, r, sess)
	}

	http.Redirect(w, r, "/clinic/pricelist", http.StatusSeeOther)
}

func clinicPricelistDeleteHandler(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if !sess.LoggedIn || sess.UserRole != "clinic" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)

	if err == nil && id >= 0 && id < len(sess.PriceList) {
		deletedService := sess.PriceList[id].Service
		sess.PriceList = append(sess.PriceList[:id], sess.PriceList[id+1:]...)
		sess.SuccessMessage = fmt.Sprintf("Услуга \"%s\" успешно удалена!", deletedService)
		saveSession(w, r, sess)
	}

	http.Redirect(w, r, "/clinic/pricelist", http.StatusSeeOther)
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

    // Handle filter submission
    if r.Method == "POST" {
        clinic := r.FormValue("clinic")
        ageGroup := r.FormValue("age_group")
        gender := r.FormValue("gender")

        // Save filters
        sess.AnalyticsFilters = AnalyticsFilters{
            Clinic:   clinic,
            AgeGroup: ageGroup,
            Gender:   gender,
        }

        // Generate filtered data
        data, diseases := getAnalyticsData(clinic, ageGroup, gender)
        sess.AnalyticsData = data
        sess.DiseaseStats = diseases

        saveSession(w, r, sess)
        http.Redirect(w, r, "/government/analytics", http.StatusSeeOther)
        return
    }

    // Initialize default data if not set
    if sess.AnalyticsData.Clinic == "" {
        data, diseases := getAnalyticsData("СтомаПрофи", "Все", "Все")
        sess.AnalyticsData = data
        sess.DiseaseStats = diseases
        sess.AnalyticsFilters = AnalyticsFilters{
            Clinic:   "СтомаПрофи",
            AgeGroup: "Все",
            Gender:   "Все",
        }
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
	http.HandleFunc("/clinic/pricelist/add", clinicPricelistAddHandler)
	http.HandleFunc("/clinic/pricelist/delete", clinicPricelistDeleteHandler)

	// Government routes
	http.HandleFunc("/government/dashboard", governmentDashboardHandler)
	http.HandleFunc("/government/analytics", governmentAnalyticsHandler)

	// API routes
	http.HandleFunc("/api/calculate", apiCalculateHandler)

	fmt.Println("🦷 DentalAI Platform starting...")
	fmt.Println("🌐 Server running on http://localhost:8080")
	fmt.Println("📱 Open your browser and navigate to http://localhost:8080")
	fmt.Println("")
	fmt.Println("🔐 Demo Credentials:")
	fmt.Println("   Patient:    username: patient    | password: demo123")
	fmt.Println("   Clinic:     username: clinic     | password: demo123")
	fmt.Println("   Government: username: government | password: demo123")
	fmt.Println("")
	fmt.Println("✨ Using layout-based templates with base.html")
	fmt.Println("🎨 Static files served from static/ directory")
	fmt.Println("📊 Demo data: 2-3 months of realistic activity")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
