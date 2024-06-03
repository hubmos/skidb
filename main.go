package main

import (
    "fmt"
	"log"
	"net/http"
	//"os"
    "github.com/joho/godotenv"
    "net/mail"
    "time"
    "os"
    "github.com/pocketbase/pocketbase/tools/mailer"
    "github.com/robfig/cron/v3"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	//"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/models"
    "github.com/pocketbase/pocketbase/forms"
)

type Leietager struct {
    Barcode string  `db:"barcode" json:"barcode"`
    Tlf     string  `db:"tlf" json:"tlf"`
    Navn    string  `db:"navn" json:"navn"`
}

func createInventoryItem(app *pocketbase.PocketBase) echo.HandlerFunc {
    return func(c echo.Context) error {

    // Find the "inventory" collection
    collection, err := app.Dao().FindCollectionByNameOrId("inventory")
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    // Create a new record for the "inventory" collection
    record := models.NewRecord(collection)

    // Initialize a new form for record upsert
    form := forms.NewRecordUpsert(app, record)

    // Load data from request. Here you need to parse the incoming request to get the data.
    // This example uses JSON body, but adjust according to your actual request format.
    var requestData map[string]any
    if err := c.Bind(&requestData); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
    }

    // Load the request data into the form
    form.LoadData(requestData)

    // Validate and submit the form (this will save the record if validation passes)
    if err := form.Submit(); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    // Return success response
    return c.JSON(http.StatusCreated, record)
	}}

func checkAndSendEmails(app *pocketbase.PocketBase, recipient string) {
    users := []Leietager{}

    err := app.Dao().DB().
        NewQuery("SELECT CONCAT(e.merke,' ',e.modell) as barcode, t.tlf, t.navn  FROM utleid t left join inventory e on e.barcode==t.barcode").
        All(&users)
    if len(users) == 0 {
        log.Println("No records to send.")
        return
    }
// Compose the email content
htmlContent := `<h1>Utleid utstyr</h1>`
htmlContent += `<table style="width: 100%; border-collapse: collapse;">`
htmlContent += `<tr style="background-color: #f8f8f8;"> <th style="border: 1px solid #ddd; padding: 8px;">Utstyr</th><th style="border: 1px solid #ddd; padding: 8px;">Leietager</th><th style="border: 1px solid #ddd; padding: 8px;">Telefon</th></tr>`

for _, record := range users {
    htmlContent += fmt.Sprintf(`<tr> <td style="border: 1px solid #ddd; padding: 8px;">%s</td><td style="border: 1px solid #ddd; padding: 8px;">%s</td><td style="border: 1px solid #ddd; padding: 8px;">%s</td></tr>`, record.Barcode, record.Navn, record.Tlf)
}

htmlContent += `</table>`

message := &mailer.Message{
    From: mail.Address{
        Address: app.Settings().Meta.SenderAddress,
        Name:    app.Settings().Meta.SenderName,
    },
    To:      []mail.Address{{Address: recipient}}, // specify recipient
    Subject: "Utleie ikke innlevert",
    HTML:    htmlContent,
    // cc, bcc, attachments and custom headers can also be added here...
}
    err = app.NewMailClient().Send(message)
    if err != nil {
        log.Println("Failed to send email:", err)
        return
    }
    log.Println("Email sent successfully")
}
func main() {
        if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
        cronSchedule := os.Getenv("CRON_SCHEDULE")
    if cronSchedule == "" {
        log.Fatalf("No CRON_SCHEDULE specified in the environment")
    }

        recipient := os.Getenv("E_MAIL")
    if recipient == "" {
        log.Fatalf("No E_MAIL  specified in the environment")
    }

	app := pocketbase.New()
    c := cron.New(cron.WithLocation(time.Local))
    _, err := c.AddFunc(cronSchedule, func() { // Run every day at 18:00
        checkAndSendEmails(app,recipient)
    })
    if err != nil {
        log.Fatal(err)
    }
    c.Start()

    defer c.Stop()
	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		//e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		e.Router.GET("/reserve/:id", func(c echo.Context) error {
			id := c.PathParam("id") // Correct method to get path parameter is Param, not PathParam.
			phonenumber := c.QueryParam("phonenumber")
			name := c.QueryParam("name")

			// Check if the record exists and its current status before proceeding
			record, err := app.Dao().FindFirstRecordByData("inventory", "barcode", id)
			if err != nil {
				return err
			}
			if record == nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Record not found"})
			}
			if record.Get("status") == 1 {
				// Assuming `utleid` is a field within the record and you're checking its value.
				// Note: You might need to cast the value of record.Data()["utleid"] to the correct type.
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Utstyret er allerede registrert som utleid."})
			}
            if record.Get("status") == 2 {
				// Assuming `utleid` is a field within the record and you're checking its value.
				// Note: You might need to cast the value of record.Data()["utleid"] to the correct type.
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Utstyret er registrert som utgått."})
			}

			record.Set("status", "1")
			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
			if err := app.Dao().SaveRecord(record); err != nil {
                log.Fatal(err)
				return err
			}
            collection, err := app.Dao().FindCollectionByNameOrId("utleid")
            if err != nil {
                return err
                }

			// Proceed to insert into "utleid" collection if not reserved
			newRecord :=models.NewRecord(collection)
			newRecord.Set("barcode", id) // Ensure you are setting the correct field names as per your DB schema.
			newRecord.Set("tlf", phonenumber)
			newRecord.Set("navn", name)
			if err := app.Dao().SaveRecord(newRecord); err != nil {
				return err
			}



			return c.JSON(http.StatusOK, map[string]string{"reserve": "1"})
		})

		e.Router.GET("/inactive/:id", func(c echo.Context) error {
			id := c.PathParam("id") // Correct method to get path parameter is Param, not PathParam.

			// Check if the record exists and its current status before proceeding
			record, err := app.Dao().FindFirstRecordByData("inventory", "barcode", id)
			if err != nil {
				return err
			}
			if record == nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Record not found"})
			}
			if record.Get("status") == 1 {
				// Assuming `utleid` is a field within the record and you're checking its value.
				// Note: You might need to cast the value of record.Data()["utleid"] to the correct type.
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Item already reserved"})
			}
            			// Optionally, update the inventory record to mark as reserved
			record.Set("status", "2")
			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, map[string]string{"reserve": "1"})
		})

		e.Router.GET("/deliver/:id", func(c echo.Context) error {
			id := c.PathParam("id")
            tilstand := c.QueryParam("tilstand")
			moreInfo := c.QueryParam("moreInfo")

			record, err := app.Dao().FindFirstRecordByData("inventory", "barcode", id)
			if err != nil {
				return err
			}
			if record != nil {
              	if record.Get("status") == 0 {
				// Assuming `utleid` is a field within the record and you're checking its value.
				// Note: You might need to cast the value of record.Data()["utleid"] to the correct type.
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Utstyr allerede registrert som innlevert."})
			}
            			// Optionally, update the inventory record to mark as reserved
              	if record.Get("status") == 2 {
				// Assuming `utleid` is a field within the record and you're checking its value.
				// Note: You might need to cast the value of record.Data()["utleid"] to the correct type.
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Utstyret er registrert som utgått."})
			}
			    record.Set("status", "0")
			    if err := app.Dao().SaveRecord(record); err != nil {
				    return err
			    }
                utleidrecord, err := app.Dao().FindFirstRecordByData("utleid", "barcode", id)
			    if err != nil {
			    	return err
			    }
                collection, err := app.Dao().FindCollectionByNameOrId("utleielogg")
                if err!= nil {
                    return err}
                utleielogg:= models.NewRecord(collection)
                utleielogg.Set("barcode", utleidrecord.Get("barcode"))
                utleielogg.Set("tilstand", tilstand)
                utleielogg.Set("leietakers_navn", utleidrecord.Get("navn"))
                utleielogg.Set("leietakers_tlf", utleidrecord.Get("tlf"))
                utleielogg.Set("tilleggsinfo", moreInfo)

                if err := app.Dao().SaveRecord(utleielogg); err != nil {
                    return err
                    }
                if err := app.Dao().DeleteRecord(utleidrecord); err != nil {
                    return err
                    }
                return c.JSON(http.StatusOK, map[string]string{"deliver": "1"})
			}
			return c.JSON(http.StatusOK, map[string]string{"deliver": "0"})
		} /* optional middlewares */)

        // Register the "create" endpoint
        e.Router.POST("/create", createInventoryItem(app))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
