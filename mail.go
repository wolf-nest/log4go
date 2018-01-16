package log4go

import (
	"fmt"
	"github.com/smartwalle/mail4go"
)

type MailWriter struct {
	level   int
	config  *mail4go.MailConfig
	subject string
	from    string
	to      []string
}

func NewMailWriter(level int) *MailWriter {
	var writer = &MailWriter{}
	writer.level = level
	return writer
}

func (this *MailWriter) SetLevel(level int) {
	this.level = level
}

func (this *MailWriter) GetLevel() int {
	return this.level
}

func (this *MailWriter) SetMailConfig(config *mail4go.MailConfig) {
	this.config = config
}

func (this *MailWriter) GetMailConfig() *mail4go.MailConfig {
	return this.config
}

func (this *MailWriter) SetSubject(subject string) {
	this.subject = subject
}

func (this *MailWriter) GetSubject() string {
	return this.subject
}

func (this *MailWriter) SetFrom(from string) {
	this.from = from
}

func (this *MailWriter) GetFrom() string {
	return this.from
}

func (this *MailWriter) SetToMailList(to ...string) {
	this.to = to
}

func (this *MailWriter) GetToMailList() []string {
	return this.to
}

func (this *MailWriter) WriteMessage(msg *LogMessage) {
	if msg == nil {
		return
	}
	if msg.level < this.level {
		return
	}

	if this.config == nil {
		return
	}

	if len(this.to) == 0 {
		return
	}

	var out = fmt.Sprintf("%s %s [%s:%d] %s", msg.header, msg.levelName, msg.file, msg.line, msg.message)

	var subject = this.GetSubject()
	if len(subject) == 0 {
		subject = msg.file
	}

	var mail = mail4go.NewTextMessage(subject, out)
	mail.To = this.to
	if len(this.from) > 0 {
		mail.From = this.from
	}

	mail4go.SendMail(this.config, mail)
}

func (this *MailWriter) Close() error {
	return nil
}
