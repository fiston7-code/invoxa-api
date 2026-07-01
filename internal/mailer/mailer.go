package mailer

import (
	"bytes"
	"embed"
	"github.com/resend/resend-go/v3"

	// Import the html/template and text/template packages. Because these share the same
	// package name ("template") we need to disambiguate them and alias them to ht and tt
	// respectively.
	ht "html/template"
	tt "text/template"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓
//
//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a mail.Client instance (used to connect to a
// SMTP server) and the sender information for your emails (the name and address you
// want the email to be from, such as "Alice Smith <alice@example.com>").
type Mailer struct {
	client *resend.Client
	sender string
}

func New(apiKey, sender string) (*Mailer, error) {
	client := resend.NewClient(apiKey)
	return &Mailer{
		client: client,
		sender: sender,
	}, nil
}

// Define a Send() method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates, and any
// dynamic data for the templates as an any parameter.
func (m *Mailer) Send(recipient string, templateFile string, data any) error {
	// Use the ParseFS() method from text/template to parse the required template file
	// from the embedded file system.
	textTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable.
	plainBody := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}
	// Use the ParseFS() method from html/template this time to parse the required template
	// file from the embedded file system.
	htmlTmpl, err := ht.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// And execute the "htmlBody" template and store the result in the htmlBody variable.
	htmlBody := new(bytes.Buffer)
	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}
	// Use the mail.NewMsg() function to initialize a new mail.Msg instance.
	// Then we use the To(), From() and Subject() methods to set the email recipient,
	// sender and subject headers, the SetBodyString() method to set the plain-text body,
	// and the AddAlternativeString() method to set the HTML body.
	params := &resend.SendEmailRequest{
		From:    m.sender,
		To:      []string{recipient},
		Subject: subject.String(),
		Text:    plainBody.String(),
		Html:    htmlBody.String(),
	}

	// Note: Template support in the Go SDK uses the Headers field
	// to pass template_id. Check the latest SDK docs for the
	// recommended approach for your version.
	// For now, this example shows the basic send pattern.
	// When templates are fully supported in the Go SDK struct,
	// you would set params.TemplateId and params.TemplateVariables.

	_, err = m.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
