package security

import (
	"errors"
	"fmt"
	"github.com/imroc/req"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/sql"
)

func (security *securityImpl) VerifyEmailServer(mail sql.ReceivedMessage) error {
	domain, err := helpers.GetHostnameFromURI(mail.URI)
	if err != nil {
		return err
	}

	security.logger.Debug(domain)
	id := fmt.Sprint(mail.ServerID)

	reqdom := domain + "/smtp2/message/verify?id=" + id + "&pass=" + mail.MVPPass
	security.logger.Debug(reqdom)
	res, err := req.Get(reqdom)
	if err != nil {
		return err
	}
	bodystring := res.String()
	if bodystring == "FAIL1" || bodystring == "FAIL2" || bodystring == "FAIL3" || bodystring == "FAIL" {
		return errors.New("failed to verify origin")
	}
	if bodystring == "OK" {
		return nil
	}
	return errors.New("invalid SMTP2 verification message")
}

func (security *securityImpl) VerifyEmailSender(mail sql.ReceivedMessage) error {
	domain, err := helpers.GetDomainFromEmail(mail.FromEmail)
	if err != nil {
		return err
	}
	security.logger.Info(domain)
	id := fmt.Sprint(mail.ServerID)

	protocol, err := security.GetProtocolFromDomain(domain)
	if err != nil {
		return err
	}

	reqdom := protocol + domain + "/smtp2/message/verify?id=" + id + "&pass=" + mail.MVPPass
	security.logger.Info(reqdom)
	res, err := req.Get(reqdom)
	if err != nil {
		return err
	}

	bodystring := res.String()
	security.logger.Debug(bodystring)
	if bodystring == "FAIL1" || bodystring == "FAIL2" || bodystring == "FAIL3" || bodystring == "FAIL" {
		return errors.New("failed to verify origin")
	}
	if bodystring == "OK" {
		return nil
	}
	return errors.New("invalid SMTP2 verification message")
}

func (security *securityImpl) VerifyMessage(mail sql.ReceivedMessage) (bool, error) {
	sendererr := security.VerifyEmailSender(mail)
	if sendererr != nil {
		return false, sendererr
	}
	servererr := security.VerifyEmailServer(mail)
	security.logger.Debug("SERVER_ERR", servererr)
	return servererr == nil, servererr
}
