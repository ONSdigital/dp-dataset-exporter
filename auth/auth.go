package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ian-kent/go-log/log"
)

func CheckServiceIdentity(ctx context.Context, zebedeeURL, serviceAuthToken string) error {
	client := rchttp.DefaultClient

	path := fmt.Sprintf("%s/identity", zebedeeURL)

	var URL *url.URL
	URL, err := url.Parse(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", serviceAuthToken)

	res, err := client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status [%d] returned from [%s]", res.StatusCode, zebedeeURL)
	}

	log.Info("dataset exporter has a valid service account", nil)
	return nil
}
