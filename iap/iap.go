// Package iap implements the ability to easily validate a receipt with apples verifyReceipt service
package iap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Documentation: https://developer.apple.com/library/ios/releasenotes/General/ValidateAppStoreReceipt/Chapters/ReceiptFields.html#//apple_ref/doc/uid/TP40010573-CH106-SW10

const (
	appleSandboxURL    string = "https://sandbox.itunes.apple.com/verifyReceipt"
	appleProductionURL string = "https://buy.itunes.apple.com/verifyReceipt"
)

type RequestDate struct {
	RequestDate    string `json:"request_date"`
	RequestDateMS  string `json:"request_date_ms"`
	RequestDatePST string `json:"request_date_pst"`
}

type PurchaseDate struct {
	PurchaseDate    string `json:"purchase_date"`
	PurchaseDateMS  string `json:"purchase_date_ms"`
	PurchaseDatePST string `json:"purchase_date_pst"`
}

type OriginalPurchaseDate struct {
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMS  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST string `json:"original_purchase_date_pst"`
}

type ExpiresDate struct {
	ExpiresDate    string `json:"expires_date"`
	ExpiresDateMS  string `json:"expires_date_ms"`
	ExpiresDatePST string `json:"expires_date_pst"`
}

type CancellationDate struct {
	CancellationDate    string `json:"cancellation_date"`
	CancellationDateMS  string `json:"cancellation_date_ms"`
	CancellationDatePST string `json:"cancellation_date_pst"`
}

type InApp struct {
	Quantity                  string `json:"quantity"`
	ProductID                 string `json:"product_id"`
	TransactionID             string `json:"transaction_id"`
	OriginalTransactionID     string `json:"original_transaction_id"`
	IsTrialPeriod             string `json:"is_trial_period"`
	AppItemID                 string `json:"app_item_id"`
	VersionExternalIdentifier string `json:"version_external_identifier"`
	WebOrderLineItemID        string `json:"web_order_line_item_id"`
	PurchaseDate
	OriginalPurchaseDate
	ExpiresDate
	CancellationDate
}

type iOS6Receipt struct {
	AppItemID                  string `json:"app_item_id"`
	BundleID                   string `json:"bid"`
	ApplicationVersion         string `json:"bvrs"`
	OriginalApplicationVersion string `json:"original_application_version"`
	OriginalTransactionID      string `json:"original_transaction_id"`
	ProductID                  string `json:"product_id"`
	Quantity                   string `json:"quantity"`
	TransactionID              string `json:"transaction_id"`
	VersionExternalIdentifier  string `json:"version_external_identifier"`
	WebOrderLineItemID         string `json:"web_order_line_item_id"`
	ExpiresDate                string `json:"expires_date_formatted"`
	ExpiresDateMS              string `json:"expires_date"`
	ExpiresDatePST             string `json:"expires_date_formatted_pst"`
	RequestDate
	PurchaseDate
	OriginalPurchaseDate
}

type Receipt struct {
	ReceiptType                string  `json:"receipt_type"`
	AdamID                     int64   `json:"adam_id"`
	AppItemID                  int64   `json:"app_item_id"`
	BundleID                   string  `json:"bundle_id"`
	ApplicationVersion         string  `json:"application_version"`
	DownloadID                 int64   `json:"download_id"`
	OriginalApplicationVersion string  `json:"original_application_version"`
	InApp                      []InApp `json:"in_app"`
	RequestDate
	OriginalPurchaseDate
}

type receiptRequestData struct {
	Receiptdata string `json:"receipt-data"`
}

// Given receiptData (base64 encoded) it tries to connect to either the sandbox (useSandbox true) or
// apples ordinary service (useSandbox false) to validate the receipt. Returns either a receipt struct or an error.
func VerifyReceipt(receiptData string, useSandbox bool) (*Receipt, error) {
	return sendReceiptToApple(receiptData, verificationURL(useSandbox))
}

// Selects the proper url to use when talking to apple based on if we should use the sandbox environment or not
func verificationURL(useSandbox bool) string {
	if useSandbox {
		return appleSandboxURL
	}
	return appleProductionURL
}

// Sends the receipt to apple, returns the receipt or an error upon completion
func sendReceiptToApple(receiptData, url string) (*Receipt, error) {
	requestData, err := json.Marshal(receiptRequestData{receiptData})
	if err != nil {
		return nil, err
	}

	toSend := bytes.NewBuffer(requestData)
	resp, err := http.Post(url, "application/json", toSend)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseData struct {
		Status         float64     `json:"status"`
		ReceiptContent iOS6Receipt `json:"receipt"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}

	if responseData.Status != 0 {
		return nil, verificationError(responseData.Status)
	}

	return responseData.ReceiptContent.toReceipt(), nil
}

func (ios6 *iOS6Receipt) toReceipt() *Receipt {
	appItemID, _ := strconv.ParseInt(ios6.AppItemID, 10, 64)
	return &Receipt{
		AppItemID:                  appItemID,
		BundleID:                   ios6.BundleID,
		ApplicationVersion:         ios6.ApplicationVersion,
		OriginalApplicationVersion: ios6.OriginalApplicationVersion,
		RequestDate:                ios6.RequestDate,
		OriginalPurchaseDate:       ios6.OriginalPurchaseDate,
		InApp: []InApp{{
			Quantity:                  ios6.Quantity,
			ProductID:                 ios6.ProductID,
			TransactionID:             ios6.TransactionID,
			OriginalTransactionID:     ios6.OriginalTransactionID,
			VersionExternalIdentifier: ios6.VersionExternalIdentifier,
			WebOrderLineItemID:        ios6.WebOrderLineItemID,
			PurchaseDate:              ios6.PurchaseDate,
			OriginalPurchaseDate:      ios6.OriginalPurchaseDate,
			ExpiresDate: ExpiresDate{
				ExpiresDate:    ios6.ExpiresDate,
				ExpiresDateMS:  ios6.ExpiresDateMS,
				ExpiresDatePST: ios6.ExpiresDatePST,
			},
		}},
	}
}

var errMsgs = map[float64]string{
	21000: "The App Store could not read the JSON object you provided.",
	21002: "The data in the receipt-data property was malformed.",
	21003: "The receipt could not be authenticated.",
	21004: "The shared secret you provided does not match the shared secret on file for your account.",
	21005: "The receipt server is not currently available.",
	21006: "This receipt is valid but the subscription has expired. When this status code is returned to your server, the receipt data is also decoded and returned as part of the response.",
	21007: "This receipt is a sandbox receipt, but it was sent to the production service for verification.",
	21008: "This receipt is a production receipt, but it was sent to the sandbox service for verification.",
}

// Generates the correct error based on a status error code
func verificationError(errCode float64) error {
	return fmt.Errorf("%f: %s", errCode, errMsgs[errCode])
}
