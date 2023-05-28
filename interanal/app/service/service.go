package service

import (
	"log"

	"github.com/andrpech/ses-gen-tech/interanal/app/service/fs"
	"github.com/andrpech/ses-gen-tech/interanal/app/service/mail"
	"github.com/andrpech/ses-gen-tech/interanal/app/service/rate"
	"github.com/andrpech/ses-gen-tech/util"
)

type Service struct {
	RateService *rate.RateService
	EmailSender mail.EmailSender
	FileService fs.FileService
}

func New() *Service {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	return &Service{
		RateService: rate.New(),
		EmailSender: mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword),
		FileService: fs.NewFS(),
	}
}

func (s *Service) GetBinanceRate() (float64, error) {
	return s.RateService.GetBinanceRate()
}

func (s *Service) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	return s.EmailSender.SendEmail(subject, content, to, cc, bcc, attachFiles)
}

func (s *Service) ReadJson(filePath string) ([]fs.Email, error) {
	return s.FileService.ReadJson(filePath)
}

func (s *Service) WriteJson(filePath string, emails []fs.Email) error {
	return s.FileService.WriteJson(filePath, emails)
}
