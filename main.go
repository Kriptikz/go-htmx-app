package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const port = "10000"

type Contact struct {
	Id    string
	First string
	Last  string
	Phone string
	Email string
}

type NewContact struct {
	Contact
	Errors map[string]string
}

type ContactList struct {
	Contacts []Contact
	lastId   int
}

var contactsList = ContactList{
	[]Contact{
		{"0", "John", "Doe", "0234567890", "john.doe@email.com"},
		{"1", "Jane", "Doe", "1234567890", "jane.doe@email.com"},
		{"2", "Jim", "Beam", "2234567890", "jim.beam@gmail.com"},
		{"3", "Jack", "Daniels", "3234567890", "jack.daniels@gmail.com"},
		{"4", "Johnny", "Walker", "4234567890", "johnny.walker@yahoo.com"},
		{"5", "James", "Bond", "5234567890", "james.bond@yahoo.com"},
		{"6", "Jill", "Valentine", "6234567890", "jill.valentine@hotmail.com"},
		{"7", "Jerry", "Seinfeld", "7234567890", "jerry.seinfeld@gmail.com"},
		{"8", "Jessica", "Jones", "8234567890", "jessica.jones@gmail.com"},
		{"9", "Jules", "Verne", "9234567890", "jules.verne@gmail.com"},
		{"10", "Adam", "Johnson", "10234567891", "adam.johnson@gmail.com"},
		{"11", "Brian", "Miller", "11234567892", "brian.miller@gmail.com"},
		{"12", "Charles", "Brown", "12234567893", "charles.brown@gmail.com"},
		{"13", "David", "Davis", "13234567894", "david.davis@gmail.com"},
		{"14", "Edward", "Martin", "14234567895", "edward.martin@yahoo.com"},
		{"15", "Frank", "Thompson", "15234567896", "frank.thompson@yahoo.com"},
		{"16", "George", "Taylor", "16234567897", "george.taylor@hotmail.com"},
		{"17", "Henry", "Anderson", "17234567898", "henry.anderson@gmail.com"},
		{"18", "Ivan", "Thomas", "18234567899", "ivan.thomas@gmail.com"},
		{"19", "Jacob", "Jackson", "19234567900", "jacob.jackson@gmail.com"},
		{"20", "Kevin", "White", "20234567901", "kevin.white@gmail.com"},
		{"21", "Louis", "Harris", "21234567902", "louis.harris@gmail.com"},
		{"22", "Michael", "Martin", "22234567903", "michael.martin@gmail.com"},
		{"23", "Nathan", "Thompson", "23234567904", "nathan.thompson@gmail.com"},
		{"24", "Oscar", "Garcia", "24234567905", "oscar.garcia@gmail.com"},
		{"25", "Paul", "Martinez", "25234567906", "paul.martinez@gmail.com"},
		{"26", "Quincy", "Robinson", "26234567907", "quincy.robinson@gmail.com"},
		{"27", "Robert", "Clark", "27234567908", "robert.clark@gmail.com"},
		{"28", "Steven", "Rodriguez", "28234567909", "steven.rodriguez@gmail.com"},
		{"29", "Thomas", "Lewis", "29234567910", "thomas.lewis@gmail.com"},
	},
	3,
}

type Archiver struct {
	Status      string  // "Waiting", "Running", "Complete"
	Progress    float64 // A number between 0 and 1
	ProgressUI  float64 // A Ui Friendly version of the Progress
	ArchiveFile string  // The path to the archived file
}

func (a *Archiver) Run(currentContacts []Contact) error {
	if a.Status != "Waiting" {
		return errors.New("Cannot start a new job unless the status is 'Waiting'")
	}

	a.Status = "Running"
	a.Progress = 0

	// Archive the contacts
	err := a.archiveContacts(currentContacts)
	if err != nil {
		a.Status = "Waiting"
		return err
	}

	a.Status = "Complete"
	a.Progress = 1
	a.ProgressUI = 100
	a.ArchiveFile = "static/archive/archived_contacts.json"

	return nil
}

func (a *Archiver) Reset() {
	a.Status = "Waiting"
	a.Progress = 0
	a.ArchiveFile = ""
}

func (a *Archiver) archiveContacts(currentContacts []Contact) error {
	totalContacts := len(currentContacts)

	// Create the file
	file, err := os.OpenFile("static/archive/archived_contacts.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the opening bracket
	_, err = file.WriteString("[")
	if err != nil {
		return err
	}

	for i, contact := range currentContacts {
		time.Sleep(200 * time.Millisecond)
		// Convert the contact to JSON
		data, err := json.Marshal(contact)
		if err != nil {
			return err
		}

		// Write the JSON data to the file
		_, err = file.Write(data)
		if err != nil {
			return err
		}

		// Write a comma after each contact
		_, err = file.WriteString(",")
		if err != nil {
			return err
		}

		// Update the progress
		a.Progress = float64(i+1) / float64(totalContacts)
		a.ProgressUI = a.Progress * 100
	}

	// Seek back one byte to overwrite the last comma
	_, err = file.Seek(-1, io.SeekEnd)
	if err != nil {
		return err
	}

	// Write the closing bracket
	_, err = file.WriteString("]")
	if err != nil {
		return err
	}

	a.ArchiveFile = "static/archive/archived_contacts.json"
	return nil
}

func (a *Archiver) DeleteArchive() error {
	// Delete the archive file
	err := os.Remove(a.ArchiveFile)
	if err != nil {
		return err
	}

	// Reset the Archiver
	a.Reset()

	return nil
}

var archiver = Archiver{
	Status: "Waiting",
}

var templates = make(map[string]*template.Template)

func doesEmailExistAlready(currentContacts []Contact, email string) bool {
	// Check if email is already in use
	// Use a hashmap here if you want to get fast and fancy. Most likely it will just be an indexed column lookup in some db though.
	for _, contact := range currentContacts {
		if contact.Email == email {
			return true
		}
	}

	return false
}

func doesEmailExistAlreadyExcludingContactId(currentContacts []Contact, email string, contactId string) bool {
	// Check if email is already in use
	// Use a hashmap here if you want to get fast and fancy. Most likely it will just be an indexed column lookup in some db though.
	for _, contact := range currentContacts {
		if contact.Email == email && contact.Id != contactId {
			return true
		}
	}

	return false
}

func addContact(first string, last string, phone string, email string, currentContacts []Contact) (NewContact, error) {
	var errors = make(map[string]string)
	var newContact NewContact

	// Can add much more thorough error handling here if needed.
	if first == "" {
		errors["First"] = "First name is required"
	}
	if last == "" {
		errors["Last"] = "Last name is required"
	}
	if phone == "" {
		errors["Phone"] = "Phone number is required"
	}
	if email == "" {
		errors["Email"] = "Email is required"
	} else {
		if doesEmailExistAlready(currentContacts, email) {
			errors["Email"] = "Email is already in use"
		}
	}

	contactsList.lastId++

	newContact = NewContact{
		Contact: Contact{
			Id:    strconv.Itoa(contactsList.lastId),
			First: first,
			Last:  last,
			Phone: phone,
			Email: email,
		},
		Errors: errors,
	}

	if len(errors) == 0 {
		newContacts := append(contactsList.Contacts, newContact.Contact)
		contactsList.Contacts = newContacts

		return newContact, nil
	} else {
		return newContact, fmt.Errorf("Errors: %v", errors)
	}
}

func getContactById(contacts []Contact, id string) (Contact, error) {
	var contact Contact
	for _, c := range contacts {
		if c.Id == id {
			return c, nil
		}
	}
	return contact, fmt.Errorf("Contact with ID %s not found", id)
}

func editContactById(contactsList *ContactList, id string, first string, last string, phone string, email string) (NewContact, error) {
	var contactErrors = make(map[string]string)
	var editedContact NewContact

	if doesEmailExistAlreadyExcludingContactId(contactsList.Contacts, email, id) {
		contactErrors["Email"] = "Email is already in use"
	}

	// Find the contact by id
	for i, c := range contactsList.Contacts {
		if c.Id == id {
			// Update the contact fields
			contactsList.Contacts[i].First = first
			contactsList.Contacts[i].Last = last
			contactsList.Contacts[i].Phone = phone
			contactsList.Contacts[i].Email = email

			// Create the NewContact
			editedContact = NewContact{
				Contact: contactsList.Contacts[i],
				Errors:  contactErrors,
			}

			var newError error
			if len(contactErrors) > 0 {
				newError = errors.New("Failed to save contact")
			}
			return editedContact, newError
		}
	}

	return editedContact, fmt.Errorf("Contact with ID %s not found", id)
}

func deleteContactById(contactsList *ContactList, id string) error {
	for i, c := range contactsList.Contacts {
		if c.Id == id {
			// Delete the contact from the list
			contactsList.Contacts = append(contactsList.Contacts[:i], contactsList.Contacts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Contact with ID %s not found", id)
}

func filterContacts(contacts []Contact, query string) []Contact {
	var filteredContacts []Contact
	if query == "" {
		return contacts
	} else {
		for _, contact := range contacts {
			if strings.Contains(contact.First, query) ||
				strings.Contains(contact.Last, query) ||
				strings.Contains(contact.Phone, query) ||
				strings.Contains(contact.Email, query) {
				filteredContacts = append(filteredContacts, contact)
			}
		}

		return filteredContacts
	}
}

func addNewContact(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when parsing form: %v", err)
		return
	}

	first := r.FormValue("first_name")
	last := r.FormValue("last_name")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	newContact, addContactErr := addContact(first, last, phone, email, contactsList.Contacts)
	if addContactErr != nil {
		err = templates["new"].ExecuteTemplate(w, "layout.html", newContact)
		if err != nil {
			fmt.Println("Error when executing template", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// The book includes a flash message here, but there isn't something similar in Go's standard lib
		// Can add your own implementation for flash, or use a third-party library
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
	}
}

func editContactEmailCheck(w http.ResponseWriter, r *http.Request, contactId string) {
	// Parse the query parameters from the request
	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Email parameter is missing")
		return
	}

	// Check if the email already exists
	if doesEmailExistAlreadyExcludingContactId(contactsList.Contacts, email, contactId) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Email is already in use")
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, " ")
	}
}

func editContactGet(w http.ResponseWriter, r *http.Request, contactId string) {
	contact, err := getContactById(contactsList.Contacts, contactId)
	if err != nil {
		fmt.Println("Error when getting contact by ID", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	newContact := NewContact{
		Contact: contact,
		Errors:  make(map[string]string),
	}

	err = templates["edit"].ExecuteTemplate(w, "layout.html", newContact)
	if err != nil {
		fmt.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func editContactPost(w http.ResponseWriter, r *http.Request, contactId string) {
	first := r.FormValue("first_name")
	last := r.FormValue("last_name")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	editedContact, err := editContactById(&contactsList, contactId, first, last, phone, email)
	if err != nil {
		fmt.Println("Error when editing contact by ID: ", err)
		err = templates["edit"].ExecuteTemplate(w, "layout.html", editedContact)
		if err != nil {
			fmt.Println("Error when executing template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/contacts/"+contactId, http.StatusSeeOther)
	return
}

func deleteContact(w http.ResponseWriter, r *http.Request, contactId string) {
	deleteContactById(&contactsList, contactId)

	hxTrigger := r.Header.Get("HX-Trigger")
	if hxTrigger == "delete-btn" {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
	} else {
		fmt.Fprintf(w, "")
	}
}

func getContactsCount(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(4 * time.Second)
	contactsCount := len(contactsList.Contacts)
	fmt.Fprintf(w, "(%d total Contacts)", contactsCount)
}

func newContact(w http.ResponseWriter, r *http.Request) {
	contact := NewContact{
		Contact: Contact{
			First: "",
			Last:  "",
			Phone: "",
			Email: "",
		},
		Errors: make(map[string]string),
	}

	err := templates["new"].ExecuteTemplate(w, "layout.html", contact)
	if err != nil {
		fmt.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func deleteContacts(w http.ResponseWriter, r *http.Request) {
	/*
		  Unfortunately Go does not automatically parse the body of a request into the form
			for a Delete method call. So we need to do it ourselves...
	*/
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when reading request body: %v", err)
		return
	}
	defer r.Body.Close()

	// Parse the form data from the request body
	values, err := url.ParseQuery(string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when parsing form data: %v", err)
		return
	}

	// Extract the "selected_contact_ids" from the parsed form data
	selectedContactIds := values["selected_contact_ids"]

	// Log the selected contact ids
	for _, id := range selectedContactIds {
		fmt.Println("Selected contact id for deletion: ", id)
	}

	// Delete each of those ids from the contactsList
	for _, id := range selectedContactIds {
		err := deleteContactById(&contactsList, id)
		if err != nil {
			fmt.Println("Error when deleting contact by ID", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Set up the data for the contacts template
	data := struct {
		Q               string
		Contacts        []Contact
		ContactsLength  int
		PreviousPage    int
		NextPage        int
		CurrentPage     int
		CurrentArchiver Archiver
	}{
		Q:               "",
		Contacts:        contactsList.Contacts,
		ContactsLength:  len(contactsList.Contacts),
		PreviousPage:    0,
		NextPage:        0,
		CurrentPage:     1,
		CurrentArchiver: archiver,
	}

	// Execute the contacts template
	err = templates["contacts"].ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		fmt.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func contacts(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when parsing form: %v", err)
		return
	}

	searchQuery := r.FormValue("q")
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 1
	}

	// Assuming you have a constant for the number of contacts per page
	const contactsPerPage = 10

	if page <= 0 {
		page = 1
	}
	if page > (len(contactsList.Contacts)/contactsPerPage)+1 {
		page = (len(contactsList.Contacts) / contactsPerPage) + 1
	}

	// Calculate the start and end indices for slicing the contacts slice
	start := (page - 1) * contactsPerPage
	end := start + contactsPerPage

	// Slice the contacts slice
	filteredContacts := filterContacts(contactsList.Contacts, searchQuery)

	// Check if end is greater than the length of the slice
	if end > len(filteredContacts) {
		end = len(filteredContacts)
	}
	pagedContacts := filteredContacts[start:end]

	// Calculate the previous and next page numbers
	previousPage := page - 1
	if previousPage < 1 {
		previousPage = 1
	}
	nextPage := page + 1

	// Check for 'HX-Trigger' header
	hxTrigger := r.Header.Get("HX-Trigger")

	// Decide which template to execute based on 'HX-Trigger' header value
	if hxTrigger == "search" {
		// This part was tricky. Needed to use "rows" not "rows.html"
		err = templates["contacts"].ExecuteTemplate(w, "rows", struct {
			Contacts       []Contact
			ContactsLength int
			NextPage       int
			WasSearched    bool
		}{
			Contacts:       pagedContacts,
			ContactsLength: len(pagedContacts),
			NextPage:       nextPage,
			WasSearched:    true,
		})
	} else {
		err = templates["contacts"].ExecuteTemplate(w, "layout.html", struct {
			Q               string
			Contacts        []Contact
			ContactsLength  int
			PreviousPage    int
			NextPage        int
			CurrentPage     int
			CurrentArchiver Archiver
			WasSearched     bool
		}{
			Q:               searchQuery,
			Contacts:        pagedContacts,
			ContactsLength:  len(pagedContacts),
			PreviousPage:    previousPage,
			NextPage:        nextPage,
			CurrentPage:     page,
			CurrentArchiver: archiver,
			WasSearched:     false,
		})
	}

	if err != nil {
		fmt.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func showContact(w http.ResponseWriter, r *http.Request, contactId string) {
	contact, err := getContactById(contactsList.Contacts, contactId)
	if err != nil {
		fmt.Println("Error when getting contact by ID", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = templates["show"].ExecuteTemplate(w, "layout.html", contact)
	if err != nil {
		fmt.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleDeleteArchiveContacts(w http.ResponseWriter, r *http.Request) {
	archiver.Reset()
	err := templates["contacts"].ExecuteTemplate(w, "archive_ui", archiver)
	if err != nil {
		log.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleGetArchiveContacts(w http.ResponseWriter, r *http.Request) {
	err := templates["contacts"].ExecuteTemplate(w, "archive_ui", archiver)
	if err != nil {
		log.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleArchiveContacts(w http.ResponseWriter, r *http.Request) {
	// Start archiver.Run in a separate goroutine
	go func() {
		err := archiver.Run(contactsList.Contacts)
		if err != nil {
			log.Println("Error when running archiver: ", err)
		}
	}()

	// sleep here for archiver to update
	time.Sleep(10 * time.Millisecond)

	// Execute the contacts template but only use the "archiver_ui" and pass on the current archiver
	err := templates["contacts"].ExecuteTemplate(w, "archive_ui", archiver)
	if err != nil {
		log.Println("Error when executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}

// Routes
func handler(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(3 * time.Second)
	switch {
	case r.URL.Path == "/":
		http.Redirect(w, r, "/contacts", http.StatusMovedPermanently)
	case r.URL.Path == "/ping":
		pong(w, r)
	case r.URL.Path == "/contacts":
		switch r.Method {
		case http.MethodGet:
			contacts(w, r)
		case http.MethodDelete:
			deleteContacts(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	case strings.HasPrefix(r.URL.Path, "/contacts/"):
		contactId := strings.TrimPrefix(r.URL.Path, "/contacts/")
		switch {
		case strings.HasSuffix(contactId, "/edit"):
			contactId = strings.TrimSuffix(contactId, "/edit")
			switch r.Method {
			case http.MethodGet:
				editContactGet(w, r, contactId)
			case http.MethodPost:
				editContactPost(w, r, contactId)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case strings.HasSuffix(contactId, "/email"):
			contactId = strings.TrimSuffix(contactId, "/email")
			switch r.Method {
			case http.MethodGet:
				editContactEmailCheck(w, r, contactId)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case contactId == "archive":
			switch r.Method {
			case http.MethodPost:
				handleArchiveContacts(w, r)
			case http.MethodGet:
				handleGetArchiveContacts(w, r)
			case http.MethodDelete:
				handleDeleteArchiveContacts(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case contactId == "count":
			switch r.Method {
			case http.MethodGet:
				getContactsCount(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case contactId == "new":
			switch r.Method {
			case http.MethodGet:
				newContact(w, r)
			case http.MethodPost:
				addNewContact(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case contactId != "":
			switch r.Method {
			case http.MethodGet:
				showContact(w, r, contactId)
			case http.MethodDelete:
				deleteContact(w, r, contactId)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			http.Redirect(w, r, "/contacts", http.StatusMovedPermanently)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// initialize the templates here instead of in each request handler
func initTemplates() (err error) {
	/*
		We can't use parse glob here due to the way we are using layout.html as the base as from the book.
		When we execute the template, we use layout.html rather than the one with the templated content.
	*/

	/*
		To use parse glob, we would need to reverse this templating.
		Instead of layout.html we would have something like header.html and footer.html.
		Which we would then use as templated content into the actual page.
	*/

	templates["new"], err = template.ParseFiles("layout.html", "new.html")
	if err != nil {
		return err
	}
	templates["edit"], err = template.ParseFiles("layout.html", "edit.html")
	if err != nil {
		return err
	}
	templates["show"], err = template.ParseFiles("layout.html", "show.html")
	if err != nil {
		return err
	}
	templates["contacts"], err = template.ParseFiles("layout.html", "index.html", "rows.html", "archive_ui.html")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Parsing Templates...")
	err := initTemplates()
	if err != nil {
		log.Println("An error occurred parsing templates: ", err)
		return
	}

	log.Println("Parsed.")

	log.Println("Starting server on port " + port)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}
