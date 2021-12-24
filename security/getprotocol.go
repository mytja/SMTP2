package security

import "github.com/imroc/req"

func (security *securityImpl) GetProtocolFromDomain(todomain string) (string, error) {
	// Ping HEAD request to see which protocol does server use
	var protocol = "https://"
	_, err := req.Head("https://" + todomain + "/smtp2")
	if err != nil {
		security.logger.Debugw("failed to request to domain using HTTPS", "err", err)
		_, err := req.Head("http://" + todomain + "/smtp2")
		if err != nil {
			security.logger.Debugw("failed to request to domain using HTTP", "err", err)
			return "", err
		}
		security.logger.Debugw("successfully requested to domain using HTTP", "err", err)
		protocol = "http://"
	}
	return protocol, nil
}
