package security

import (
	"errors"
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/sql"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

func VerifyEmailServer(mail sql.ReceivedMessage, logger *zap.SugaredLogger) error {
	domain, err := helpers.GetHostnameFromURI(mail.URI)
	if err != nil {
		return err
	}
	logger.Info(domain)
	id := fmt.Sprint(mail.ServerID)
	reqdom := domain + "/smtp2/message/verify?id=" + id + "&pass=" + mail.MVPPass
	logger.Info(reqdom)
	res, err := http.Get(reqdom)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	bodystring := helpers.BytearrayToString(body)
	if bodystring == "FAIL1" || bodystring == "FAIL2" || bodystring == "FAIL3" || bodystring == "FAIL" {
		return errors.New("failed to verify origin")
	}
	if bodystring == "OK" {
		return nil
	}
	return errors.New("invalid SMTP2 verification message")
}

func VerifyEmailSender(mail sql.ReceivedMessage, logger *zap.SugaredLogger) error {
	domain, err := helpers.GetDomainFromEmail(mail.FromEmail)
	if err != nil {
		return err
	}
	logger.Info(domain)
	id := fmt.Sprint(mail.ServerID)
	protocol := "http://"
	if constants.ForceHttpsForMailDomain {
		protocol = "https://"
	}
	reqdom := protocol + domain + "/smtp2/message/verify?id=" + id + "&pass=" + mail.MVPPass
	logger.Info(reqdom)
	res, err := http.Get(reqdom)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	bodystring := helpers.BytearrayToString(body)
	logger.Info(bodystring)
	if bodystring == "FAIL1" || bodystring == "FAIL2" || bodystring == "FAIL3" || bodystring == "FAIL" {
		return errors.New("failed to verify origin")
	}
	if bodystring == "OK" {
		return nil
	}
	return errors.New("invalid SMTP2 verification message")
}

func VerifyMessage(mail sql.ReceivedMessage, logger *zap.SugaredLogger) (bool, error) {
	sendererr := VerifyEmailSender(mail, logger)
	if sendererr != nil {
		return false, sendererr
	}
	servererr := VerifyEmailServer(mail, logger)
	logger.Info("SERVER_ERR", servererr)
	return servererr == nil, servererr
}
