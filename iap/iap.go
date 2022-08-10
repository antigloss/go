/*
 *
 * iap - In App Purchase
 * Copyright (C) 2015 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package iap implements the ability to easily validate a receipt with Apple's VerifyReceipt service (compatible with iOS6 and iOS7 response).
// Documentation: https://developer.apple.com/library/ios/releasenotes/General/ValidateAppStoreReceipt/Chapters/ReceiptFields.html#//apple_ref/doc/uid/TP40010573-CH106-SW10
package iap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	appleSandboxURL    string = "https://sandbox.itunes.apple.com/verifyReceipt"
	appleProductionURL string = "https://buy.itunes.apple.com/verifyReceipt"
)

type requestDate struct {
	RequestDate    string `json:"request_date"`
	RequestDateMS  string `json:"request_date_ms"`
	RequestDatePST string `json:"request_date_pst"`
}

type purchaseDate struct {
	PurchaseDate    string `json:"purchase_date"`
	PurchaseDateMS  string `json:"purchase_date_ms"`
	PurchaseDatePST string `json:"purchase_date_pst"`
}

type originalPurchaseDate struct {
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMS  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST string `json:"original_purchase_date_pst"`
}

type expiresDate struct {
	ExpiresDate    string `json:"expires_date"`
	ExpiresDateMS  string `json:"expires_date_ms"`
	ExpiresDatePST string `json:"expires_date_pst"`
}

type cancellationDate struct {
	CancellationDate    string `json:"cancellation_date"`
	CancellationDateMS  string `json:"cancellation_date_ms"`
	CancellationDatePST string `json:"cancellation_date_pst"`
}

type inApp struct {
	Quantity                  string `json:"quantity"`
	ProductID                 string `json:"product_id"`
	TransactionID             string `json:"transaction_id"`
	OriginalTransactionID     string `json:"original_transaction_id"`
	IsTrialPeriod             string `json:"is_trial_period"`
	AppItemID                 string `json:"app_item_id"`
	VersionExternalIdentifier string `json:"version_external_identifier"`
	WebOrderLineItemID        string `json:"web_order_line_item_id"`
	purchaseDate
	originalPurchaseDate
	expiresDate
	cancellationDate
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
	requestDate
	purchaseDate
	originalPurchaseDate
}

type Receipt struct {
	ReceiptType                string  `json:"receipt_type"`
	AdamID                     int64   `json:"adam_id"`
	AppItemID                  int64   `json:"app_item_id"`
	BundleID                   string  `json:"bundle_id"`
	ApplicationVersion         string  `json:"application_version"`
	DownloadID                 int64   `json:"download_id"`
	OriginalApplicationVersion string  `json:"original_application_version"`
	InApp                      []inApp `json:"in_app"`
	requestDate
	originalPurchaseDate
}

type receiptRequestData struct {
	ReceiptData string `json:"receipt-data"`
}

type iOS6ResponseData struct {
	Status         int         `json:"status"`
	ReceiptContent iOS6Receipt `json:"receipt"`
}

type receiptResponseData struct {
	Status         int     `json:"status"`
	ReceiptContent Receipt `json:"receipt"`
}

// VerifyReceipt tries to connect to either the sandbox (useSandbox true) or
// Apple's ordinary service (useSandbox false) to validate the base64-encoded receipt (receiptData).
// Returns either a Receipt struct or an error.
func VerifyReceipt(receiptData string, useSandbox bool) (*Receipt, error) {
	if !useSandbox {
		return sendReceiptToApple(receiptData, appleProductionURL)
	}
	return sendReceiptToApple(receiptData, appleSandboxURL)
}

// Sends the receipt to Apple, returns the Receipt or an error upon completion.
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// for iOS7
	var responseData receiptResponseData
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}
	if responseData.Status != 0 {
		return nil, verificationError(responseData.Status)
	}
	if len(responseData.ReceiptContent.BundleID) > 0 {
		return &responseData.ReceiptContent, nil
	}

	// for iOS6
	var ios6ResponseData iOS6ResponseData
	err = json.Unmarshal(body, &ios6ResponseData)
	if err != nil {
		return nil, err
	}
	return ios6ResponseData.ReceiptContent.toReceipt(), nil
}

// Turns an iOS6Receipt into a Receipt struct
func (ios6 *iOS6Receipt) toReceipt() *Receipt {
	appItemID, _ := strconv.ParseInt(ios6.AppItemID, 10, 64)
	return &Receipt{
		AppItemID:                  appItemID,
		BundleID:                   ios6.BundleID,
		ApplicationVersion:         ios6.ApplicationVersion,
		OriginalApplicationVersion: ios6.OriginalApplicationVersion,
		requestDate:                ios6.requestDate,
		originalPurchaseDate:       ios6.originalPurchaseDate,
		InApp: []inApp{{
			Quantity:                  ios6.Quantity,
			ProductID:                 ios6.ProductID,
			TransactionID:             ios6.TransactionID,
			OriginalTransactionID:     ios6.OriginalTransactionID,
			VersionExternalIdentifier: ios6.VersionExternalIdentifier,
			WebOrderLineItemID:        ios6.WebOrderLineItemID,
			purchaseDate:              ios6.purchaseDate,
			originalPurchaseDate:      ios6.originalPurchaseDate,
			expiresDate: expiresDate{
				ExpiresDate:    ios6.ExpiresDate,
				ExpiresDateMS:  ios6.ExpiresDateMS,
				ExpiresDatePST: ios6.ExpiresDatePST,
			},
		}},
	}
}

// Maps error codes to error messages.
var errMsgs = map[int]string{
	21000: "The App Store could not read the JSON object you provided.",
	21002: "The data in the receipt-data property was malformed.",
	21003: "The receipt could not be authenticated.",
	21004: "The shared secret you provided does not match the shared secret on file for your account.",
	21005: "The receipt server is not currently available.",
	21006: "This receipt is valid but the subscription has expired. When this status code is returned to your server, the receipt data is also decoded and returned as part of the response.",
	21007: "This receipt is a sandbox receipt, but it was sent to the production service for verification.",
	21008: "This receipt is a production receipt, but it was sent to the sandbox service for verification.",
}

// Generates the correct error based on a status error code.
func verificationError(errCode int) error {
	return fmt.Errorf("%d: %s", errCode, errMsgs[errCode])
}
