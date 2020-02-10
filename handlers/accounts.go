/*
 *    Copyright 2020 bitfly gmbh
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package handlers

import (
	"coda-explorer/db"
	"coda-explorer/templates"
	"coda-explorer/types"
	"coda-explorer/version"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

var accountsTemplate = template.Must(template.New("blocks").Funcs(templates.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/accounts.html"))

// Blocks will return information about blocks using a go template
func Accounts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "coda explorer",
			Description: "",
			Path:        "",
		},
		ShowSyncingMessage: false,
		Active:             "accounts",
		Data:               nil,
		Version:            version.Version,
	}

	err := accountsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// BlocksData will return information about blocks
func AccountsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseInt(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseInt(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var accountsCount int64

	err = db.DB.Get(&accountsCount, "SELECT COUNT(*) FROM accounts")
	if err != nil {
		logger.Errorf("error retrieving accounts count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var accounts []*types.Account

	err = db.DB.Select(&accounts, `SELECT *
										FROM accounts 
										ORDER BY accounts.balance DESC LIMIT $1 OFFSET $2`, length, start)

	if err != nil {
		logger.Errorf("error retrieving accounts data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(accounts))
	for i, b := range accounts {
		tableData[i] = []interface{}{
			b.PublicKey,
			b.Balance,
			b.FirstSeen.Unix(),
			b.LastSeen.Unix(),
			b.BlocksProposed,
			b.SnarkJobs,
			b.TxSent,
			b.TxReceived,
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    accountsCount,
		RecordsFiltered: accountsCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}