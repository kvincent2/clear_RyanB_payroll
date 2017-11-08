package main

// go run main.go [Neat access key] [WLS access key]

import (
	"encoding/csv"
	"fmt"
	quickbooks "github.com/jinmatt/go-quickbooks.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// //Read CSV file with Ryan's payroll activity for the month
func main() {

	year, month, _ := time.Now().Date()

	ryanB_activity, err := ioutil.ReadFile(fmt.Sprintf("./csv/%d%d.csv", int(month-1), year))
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(strings.NewReader(string(ryanB_activity)))

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	ryanSalary := 0.00
	ryanTaxes := 0.00

	if !headersOk(records) {
		panic("Headers changed!")
	}

	ryanSalary = stringToFloat(records[1][5])
	ryanTaxes = (stringToFloat(records[1][8]) + stringToFloat(records[1][12]) + stringToFloat(records[1][14]) + stringToFloat(records[1][29]))

	neatJournalEntryLines := []quickbooks.Line{}
	neatJournalEntryLines = append(neatJournalEntryLines, createJournalEntryLine("0", "To recover Ryan B's salary from WLS", ryanSalary, "Credit", "217", "President", "", "", ""))
	neatJournalEntryLines = append(neatJournalEntryLines, createJournalEntryLine("1", "To recover Ryan B's taxes and benefits from WLS", ryanTaxes, "Credit", "222", "Executive (non Tech) Payroll Taxes &amp; Benefits", "", "", ""))

	debitLine := createJournalEntryLine("2", "To recover Ryan B's salary, taxes, and benefits from WLS", ryanTaxes+ryanSalary, "Debit", "225", "Intercompany Receivable", "Customer", "139", "WLS")
	neatJournalEntryLines = append(neatJournalEntryLines, debitLine)

	// WLSJournalEntryLines := []quickbooks.Line{}
	// WLSJournalEntryLines = append(WLSJournalEntryLines, createJournalEntryLine("0", "To recover Ryan B's salary from WLS", ryanSalary, "Debit", "93", "600.5 President", "", "", ""))
	// WLSJournalEntryLines = append(WLSJournalEntryLines, createJournalEntryLine("1", "To recover Ryan B's taxes and benefits from WLS", ryanTaxes, "Debit", "170", "600.5.1 Payroll Taxes &amp; Benefits", "", "", ""))
	// WLSJournalEntryLines = append(WLSJournalEntryLines, createJournalEntryLine("2", "To recover Ryan B's salary, taxes, and benefits from WLS", ryanTaxes+ryanSalary, "Credit", "174", "1300.7 Intercompany Payable", "", "", ""))

	// create quickbooks client
	// expects QBO_realmID_production environment variable to be set
	// expects to be passed access key from quickbooks developer dashboard
	// Post Neat's Journal Entries related to Ryan B's payroll expense
	quickbooksClient := quickbooks.NewClient(os.Getenv("QBO_realmID_production"), os.Args[1], false)

	neatJournalEntry := quickbooks.Journalentry{
		TxnDate: fmt.Sprintf("%d-%d-%d", year, int(month-1), daysIn(month-1, year)),
		Line:    neatJournalEntryLines,
	}

	JournalentryObject, err := quickbooksClient.CreateJE(neatJournalEntry)
	fmt.Print(JournalentryObject, err)

}

// //expects QBO_realmID_prod_WLS environment variable to be set
// //expects to be passed access key from quickbooks developer dashboard
// //Post WLS's Journal Entries related to Ryan B's payroll expense
// quickbooksClient := quickbooks.NewClient(os.Getenv("QBO_realmID_prod_WLS"), os.Args[4], false)

// WLSJournalEntry := quickbooks.Journalentry{
// 	TxnDate: "2017-10-30",
// 	Line:    WLSJournalEntryLines,
// }

// JournalentryObject, err := quickbooksClient.CreateJE(WLSJournalEntry)

//write function to check headers for changes
func stringToFloat(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

//How many days in each month?
func daysIn(m time.Month, year int) int {
	// This is equivalent to time.daysIn(m, year).
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return 0
}

//write function to check headers for changes
func headersOk(records [][]string) bool {
	if records[0][5] != "Regular (Amount)" {
		fmt.Print(records[0][5])
		return false
	}
	if records[0][8] != "Guideline Traditional 401(k) (Company Contribution)" {
		fmt.Print(records[0][8])
		return false
	}
	if records[0][12] != "Employee Medical Insurance (Company Contribution)" {
		fmt.Print(records[0][12])
		return false
	}
	if records[0][14] != "Dependents Medical Insurance (Company Contribution)" {
		fmt.Print(records[0][14])
		return false
	}
	if records[0][29] != "Employer Taxes" {
		fmt.Print(records[0][29])
		return false
	}
	return true
}

func createJournalEntryLine(lineID string, description string, amount float64, postingType string, accountNum string, accountName string, entityType string, entityID string, entityName string) quickbooks.Line {
	line := quickbooks.Line{
		LineID:      lineID,
		Description: description,
		Amount:      amount,
		DetailType:  "JournalEntryLineDetail",
		JournalEntryLineDetail: &quickbooks.JournalEntryLineDetail{
			PostingType: postingType,
			AccountRef: quickbooks.JournalEntryRef{
				Value: accountNum,
				Name:  accountName,
			},
		},
	}

	if entityType != "" {
		line.JournalEntryLineDetail.Entity = quickbooks.Entity{
			Type: entityType,
			EntityRef: quickbooks.JournalEntryRef{
				Value: entityID,
				Name:  entityName,
			},
		}
	}

	return line
}
