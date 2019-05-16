package log4go

import (
	"errors"
	"fmt"
	"github.com/smartwalle/mail4go"
	"time"
)

type MailWriter struct {
	level   int
	config  *mail4go.MailConfig
	subject string
	from    string
	to      []string
}

func NewMailWriter(level int) *MailWriter {
	var mw = &MailWriter{}
	mw.level = level
	return mw
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

func (this *MailWriter) SetToMail(to ...string) {
	this.to = to
}

func (this *MailWriter) GetToMailList() []string {
	return this.to
}

func (this *MailWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if this.config == nil {
		return -1, errors.New("邮件配置信息为空")
	}

	if len(this.to) == 0 {
		return -1, errors.New("收件人信息不能为空")
	}

	var subject = this.GetSubject()
	if len(subject) == 0 {
		subject = "Log4go"
	}

	var mail = mail4go.NewTextMessage(subject, string(p))
	mail.To = this.to
	if len(this.from) > 0 {
		mail.From = this.from
	}

	err = mail4go.SendWithConfig(this.config, mail)
	return 0, err
}

func (this *MailWriter) Close() error {
	return nil
}

func (this *MailWriter) Level() int {
	return this.level
}

func (this *MailWriter) WriteMessage(logTime time.Time, service, instance, prefix, timeStr string, level int, levelName, file string, line int, msg string) {
	fmt.Fprintf(this, "%s%s%s%s %s %s:%d %s", service, instance, prefix, timeStr, levelName, file, line, msg)
}
