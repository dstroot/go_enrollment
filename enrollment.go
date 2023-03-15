// Copyright 2015 Tax Products Group
// ----------------------------------------------------------------
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ---------------------------------------------------------------
// This program processes enrollment files
// 1) It validates the file structure and contents by comparing the
//    header to the file contents
// 2) It validates the contents of each record
// 3) It inserts the records into SQL Server
//    - We need to break up the "flat" record structure into a
//      relational format.

package main

import (
	// "bufio"
	"database/sql" // https://golang.org/pkg/database/sql/
	// https://golang.org/pkg/encoding/csv/
	"encoding/xml" // https://golang.org/pkg/encoding/xml/
	"flag"         // https://golang.org/pkg/flag/
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	// Notice that we're loading the MSSQL driver anonymously, aliasing its
	// package qualifier to _ so none of its exported names are visible
	// to our code. Under the hood, the driver registers itself as being
	// available to the database/sql package.
	_ "github.com/denisenkom/go-mssqldb" // https://github.com/denisenkom/go-mssqldb

	// _ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper" // https://github.com/spf13/viper
	// "github.com/spf13/cobra"              // https://github.com/spf13/cobra
	// "github.com/spf13/pflag"              //https://github.com/spf13/pflag
	"github.com/asaskevich/govalidator" // https://github.com/asaskevich/govalidator
)

// Reading files requires checking most calls for errors.
// This helper will streamline our error checks below.
func check(e error) {
	if e != nil {
		log.Fatal(e)
		// fmt.Println("Error: ", e)
		// panic(e)
	}
}

// Read in the config file (using Viper package)
func setupEnvironment() {
	viper.AddConfigPath("./config/")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	check(err)
	// if err != nil {
	// 	panic(fmt.Errorf("Fatal error with config file: %s \n", err))
	// }
}

func main() {

	setupEnvironment()

	// Now, let's grab any command line flags
	// Use -debug to turn on debugging
	debug := flag.Bool("debug", false, "enable debugging")
	// Once all flags are declared, call flag.Parse() to execute the command-line parsing.
	flag.Parse()

	// Finally, let's get any command line arguments
	// Note: os.Args[0] (first value in this slice) is the path
	// to the program so we skip it.
	args := os.Args[1:] // this creates a slice

	argString := ""
	spacer := ""

	// Step through the slice
	for _, arg := range args {
		argString += spacer + arg
		spacer = ", "
	}

	if *debug {
		fmt.Println("Command Line Arguments: ", argString)
	}

	/*

	   The idiomatic way to use a SQL, or SQL-like, database in Go is through the
	   database/sql package. It provides a lightweight interface to a row-oriented database.

	   The first thing to do is import the database/sql package, and a driver package.
	   You generally shouldn't use the driver package directly, although some drivers
	   encourage you to do so. (In our opinion, it's usually a bad idea.) Instead,
	   your code should only refer to database/sql. This helps avoid making your
	   code dependent on the driver, so that you can change the underlying driver
	   (and thus the database you're accessing) without changing your code. It also
	   forces you to use the Go idioms instead of ad-hoc idioms that a particular
	   driver author may have provided.

	   Step 1: Establish database connection.

	   The sql.DB performs some important tasks for you behind the scenes:
	    1. It opens and closes connections to the actual underlying database, via the driver.
	    2. It manages a pool of connections as needed.

	   These "connections" may be file handles, sockets, network connections,
	   or other ways to access the database. The sql.DB abstraction is
	   designed to keep you from worrying about how to manage concurrent
	   access to the underlying datastore. A connection is marked in-use
	   when you use it to perform a task, and then returned to the available
	   pool when it's not in use anymore. One consequence of this is that if
	   you fail to release connections back to the pool, you can cause db.SQL
	   to open a lot of connections, potentially running out of resources
	   (too many connections, too many open file handles, lack of available n
	   etwork ports, etc). We'll discuss more about this later.

	   To create a sql.DB, you use sql.Open(). This returns a *sql.DB:

	   Perhaps counter-intuitively, sql.Open() does
	   not establish any connections to the database, nor does it validate
	   driver connection parameters. Instead, it simply prepares the database
	   abstraction for later use.

	   // MySQL Example
	   db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/hello")
	   if err != nil {
	       log.Fatal(err)
	   }
	*/

	// SQL Server Example
	connString := "server=" + viper.GetString("mssql.host") + ";port=" + viper.GetString("mssql.port") + ";user id=" + viper.GetString("mssql.user") + ";password=" + viper.GetString("mssql.password") + ";database=" + viper.GetString("mssql.database")

	db, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}

	// It is idiomatic to defer db.Close() if the sql.DB should not have a
	// lifetime beyond the scope of the function.

	// Although it's idiomatic to Close() the database when you're finished
	// with it, the sql.DB object is designed to be long-lived. Don't Open()
	// and Close() databases frequently. Instead, create one sql.DB object
	// for each distinct datastore you need to access, and keep it until the
	// program is done accessing that datastore. Pass it around as needed, or
	// make it available somehow globally, but keep it open. And don't Open()
	// and Close() from a short-lived function. Instead, pass the sql.DB into
	// that short-lived function as an argument.
	defer db.Close()

	// The first actual connection to the underlying datastore will be
	// established lazily, when it's needed for the first time. If you want
	// to check right away that the database is available and accessible
	// (for example, check that you can establish a network connection and log
	// in), use db.Ping() to do that, and remember to check for errors:
	err = db.Ping()
	if err != nil {
		if *debug {
			fmt.Printf("Database Connected!\n")
			fmt.Printf("Connection: %s\n\n", connString)
		}
	}

	// // Perhaps the most basic file reading task is slurping a file’s entire contents into memory.
	// dat, err := ioutil.ReadFile("./tmp/file.txt")
	// check(err)
	// fmt.Print(string(dat))

	/* ********************************************************************

	   READ IN XML EXAMPLE

	********************************************************************* */

	// Golang has a very powerful encoding/xml package that is part of the
	// standard library. All you need to do is the create the data structures
	// that map to an XML document. Then read the XML document with the
	// xml.Unmarshal() function. Because Unmarshal uses the reflect package,
	// it can only assign to exported (upper case) fields.

	// OfficeInfo -
	type OfficeInfo struct {
		OfficeName          string `xml:"OfficeName"`
		PrimaryContactFirst string `xml:"PrimaryContactFirst"`
		PrimaryContactLast  string `xml:"PrimaryContactLast"`
		PhoneNumber         string `xml:"PhoneNumber"`
		FaxNumber           string `xml:"FaxNumber"`
		Email               string `xml:"Email"`
		Address1            string `xml:"Address1"`
		Address2            string `xml:"Address2"`
		City                string `xml:"City"`
		State               string `xml:"State"`
		Zip                 string `xml:"Zip"`
	}

	// OwnerInformation -
	type OwnerInformation struct {
		FirstName   string `xml:"FirstName"`
		LastName    string `xml:"LastName"`
		PhoneNumber string `xml:"PhoneNumber"`
		Email       string `xml:"Email"`
		Address1    string `xml:"Address1"`
		Address2    string `xml:"Address2"`
		City        string `xml:"City"`
		State       string `xml:"State"`
		Zip         string `xml:"Zip"`
		SSN         string `xml:"SSN"`
		DateOfBirth string `xml:"DateOfBirth"`
	}

	// EFINOwnerInfo -
	type EFINOwnerInfo struct {
		FirstName   string `xml:"FirstName"`
		LastName    string `xml:"LastName"`
		PhoneNumber string `xml:"PhoneNumber"`
		Email       string `xml:"Email"`
		Address1    string `xml:"Address1"`
		Address2    string `xml:"Address2"`
		City        string `xml:"City"`
		State       string `xml:"State"`
		Zip         string `xml:"Zip"`
		SSN         string `xml:"SSN"`
		DateOfBirth string `xml:"DateOfBirth"`
	}

	// PriorYearInfo -
	type PriorYearInfo struct {
		Bank                  string `xml:"Bank"`
		ClientOfYoursLastYear bool   `xml:"ClientOfYoursLastYear"`
	}

	// Enrollment - Enrollment record
	type Enrollment struct {
		MasterEfin       string           `xml:"MasterEfin"`
		EFIN             string           `xml:"EFIN"`
		TransmitterID    string           `xml:"TransmitterId"`
		ProcessingYear   string           `xml:"ProcessingYear"`
		OfficeInfo       OfficeInfo       `xml:"OfficeInfo"`
		OwnerInformation OwnerInformation `xml:"OwnerInformation"`
		EFINOwnerInfo    EFINOwnerInfo    `xml:"EFINOwnerInfo"`
		PriorYearInfo    PriorYearInfo    `xml:"PriorYearInfo"`
		TransactionDate  string           `xml:"TransactionDate"`
	}

	// EnrollmentCollection - Full enrollment collection
	type EnrollmentCollection struct {
		XMLName        xml.Name     `xml:"EnrollmentCollection"`
		EnrollmentList []Enrollment `xml:"Enrollment"`
	}

	// Golang has a very powerful encoding/xml package that is part of the
	// standard library. All you need to do is the create the data structures
	// that map to an XML document. Then read the XML document with the
	// xml.Unmarshal() function. Because Unmarshal uses the reflect package,
	// it can only assign to exported (upper case) fields.

	// OfficeInfo -
	type ValidOfficeInfo struct {
		OfficeName          string `valid:"alphanum,required"`
		PrimaryContactFirst string `valid:"alphanum,required"`
		PrimaryContactLast  string `valid:"alphanum,required"`
		PhoneNumber         string `valid:"-"`
		FaxNumber           string `valid:"-"`
		Email               string `valid:"email,required"`
		Address1            string `valid:"alphanum,required"`
		Address2            string `valid:"-"`
		City                string `valid:"alphanum,required"`
		State               string `valid:"length(2|2)"`
		Zip                 string `valid:"alphanum,required"`
	}

	// OwnerInformation -
	type ValidOwnerInformation struct {
		FirstName   string `valid:"alphanum,required"`
		LastName    string `valid:"alphanum,required"`
		PhoneNumber string `valid:"alphanum,required"`
		Email       string `valid:"Email"`
		Address1    string `valid:"alphanum,required"`
		Address2    string `valid:"-"`
		City        string `valid:"alphanum,required"`
		State       string `valid:"length(2|2)"`
		Zip         string `valid:"alphanum,required"`
		SSN         string `valid:"ssn"`
		DateOfBirth string `valid:"-"`
	}

	// EFINOwnerInfo -
	type ValidEFINOwnerInfo struct {
		FirstName   string `valid:"-"`
		LastName    string `valid:"-"`
		PhoneNumber string `valid:"-"`
		Email       string `valid:"-"`
		Address1    string `valid:"-"`
		Address2    string `valid:"-"`
		City        string `valid:"-"`
		State       string `valid:"-"`
		Zip         string `valid:"-"`
		SSN         string `valid:"-"`
		DateOfBirth string `valid:"-"`
	}

	// PriorYearInfo -
	type ValidPriorYearInfo struct {
		Bank                  string `valid:"-"`
		ClientOfYoursLastYear bool   `valid:"-"`
	}

	// Enrollment - Enrollment record
	type ValidEnrollment struct {
		MasterEfin       string `valid:"numeric,required"`
		EFIN             string `valid:"numeric,required"`
		TransmitterID    string `valid:"numeric,required"`
		ProcessingYear   string `valid:"numeric,required"`
		OfficeInfo       OfficeInfo
		OwnerInformation OwnerInformation
		EFINOwnerInfo    EFINOwnerInfo
		PriorYearInfo    PriorYearInfo
		TransactionDate  string `valid:"-"`
	}

	// Read the XML File
	xmlFile, err := os.Open("./examples/EROEnrollmentRecords.xml")
	check(err)
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)
	v := EnrollmentCollection{}

	// Unmarshal parses the XML-encoded data and stores the result in the
	// value pointed to by v, which must be an arbitrary struct, slice, or
	// string. Well-formed data that does not fit into v is discarded.

	err = xml.Unmarshal(b, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	// Lets view some of the data
	for _, Enrollment := range v.EnrollmentList {
		// fmt.Printf("\t%s\n\n", Enrollment)
		fmt.Printf("Tax Year: %q\n", Enrollment.ProcessingYear)
		fmt.Printf("EFIN: %q\n", Enrollment.EFIN)
		fmt.Printf("Company Name: %q\n", Enrollment.OfficeInfo.OfficeName)
		fmt.Printf("Date: %q\n", Enrollment.TransactionDate)

		// Convert date string to time value
		t, err := time.Parse(time.RFC3339, Enrollment.TransactionDate+"Z")
		if err != nil {
			println("error: " + err.Error())
		}
		fmt.Println(t)

		// let's validate some of the data
		println(govalidator.IsUpperCase(`Enrollment.OfficeInfo.OfficeName`))
		//
		//
		//
		// ValidEnrollment = Enrollment
		// result, err := govalidator.ValidateStruct(ValidEnrollment)
		// if err != nil {
		// 	println("error: " + err.Error())
		// }
		// println(result)
		// fmt.Printf("Tax Year: %q\n", ValidEnrollment.ProcessingYear)
		// fmt.Printf("EFIN: %q\n", ValidEnrollment.EFIN)
		// fmt.Printf("Company Name: %q\n", ValidEnrollment.OfficeInfo.OfficeName)
		// fmt.Printf("Date: %q\n", ValidEnrollment.TransactionDate)

		// Let's insert into SQL Server
		stmt, err := db.Prepare("INSERT INTO ero(EFIN,COMPANY,TAX_YEAR,RECEIVED_DATE) VALUES(?,?,?,?)")
		check(err)

		res, err := stmt.Exec(Enrollment.EFIN, Enrollment.OfficeInfo.OfficeName, 2016, t)
		check(err)

		// lastId, err := res.LastInsertId()
		// check(err)

		rowCnt, err := res.RowsAffected()
		check(err)

		// log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
		log.Printf("Insert Successful, Rows affected = %d\n\n", rowCnt)
	}

	/* ********************************************************************

	   END XML EXAMPLE

	********************************************************************* */

	// // You’ll often want more control over how and what parts of a file are read.
	// // For these tasks, start by Opening a file to obtain an os.File value.
	// file, err := os.Open("./examples/file.txt")
	// check(err)
	// // automatically call Close() at the end of current method
	// defer file.Close()
	//
	// // setup CSV parsing
	// reader := csv.NewReader(file)
	// reader.Comma = ','
	// lineCount := 0
	//
	// for {
	//
	// 	// read just one record, but we could ReadAll() as well
	// 	record, err := reader.Read()
	// 	if err == io.EOF {
	// 		break
	// 	} else {
	// 		check(err)
	// 	}
	//
	// 	// record is an array of string so is directly printable
	// 	fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields:")
	//
	// 	// and we can iterate on top of that through each field
	// 	for i := 0; i < len(record); i++ {
	// 		fmt.Println(record[i])
	// 	}
	//
	// 	// Querying the database:
	// 	//
	// 	// We're using db.Query() to send the query to the database. We check the error, as usual.
	// 	// We defer rows.Close(). This is very important; more on that in a moment.
	// 	// We iterate over the rows with rows.Next().
	// 	// We read the columns in each row into variables with rows.Scan().
	// 	// We check for errors after we're done iterating over the rows.
	//
	// 	// var (
	// 	//     id int
	// 	//     name string
	// 	// )
	// 	// rows, err := db.Query("select id, name from users where id = ?", 1)
	// 	// check(err)
	// 	// defer rows.Close()
	// 	// for rows.Next() {
	// 	//     err := rows.Scan(&id, &name)
	// 	//     check(err)
	// 	//     log.Println(id, name)
	// 	// }
	// 	// err = rows.Err()
	// 	// check(err)
	//
	// 	// As mentioned previously, you should only use Query functions to execute
	// 	// queries -- statements that return rows. Use Exec, preferably with a
	// 	// prepared statement, to accomplish an INSERT, UPDATE, DELETE, or other
	// 	// statement that doesn't return rows. The following example shows how to
	// 	// insert a row:
	//
	// 	// stmt, err := db.Prepare("INSERT INTO users(name) VALUES(?)")
	// 	// check(err)
	// 	//
	// 	// res, err := stmt.Exec("Dolly")
	// 	// check(err)
	// 	//
	// 	// lastId, err := res.LastInsertId()
	// 	// check(err)
	// 	//
	// 	// rowCnt, err := res.RowsAffected()
	// 	// check(err)
	// 	//
	// 	// log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
	//
	// 	fmt.Println()
	// 	lineCount++
	// }

	// // Read some bytes from the beginning of the file. Allow up to 5 to be read but also note how many actually were read.
	// b1 := make([]byte, 5)
	// n1, err := file.Read(b1)
	// check(err)
	// fmt.Printf("%d bytes: %s\n", n1, string(b1))
	//
	// // You can also Seek to a known location in the file and Read from there.
	// o2, err := file.Seek(6, 0)
	// check(err)
	// b2 := make([]byte, 2)
	// n2, err := file.Read(b2)
	// check(err)
	// fmt.Printf("%d bytes @ %d: %s\n", n2, o2, string(b2))
	//
	// // The io package provides some functions that may be helpful for file reading. For example, reads like the ones above can be more robustly implemented with ReadAtLeast.
	// o3, err := file.Seek(6, 0)
	// check(err)
	// b3 := make([]byte, 2)
	// n3, err := io.ReadAtLeast(f, b3, 2)
	// check(err)
	// fmt.Printf("%d bytes @ %d: %s\n", n3, o3, string(b3))
	//
	// //There is no built-in rewind, but Seek(0, 0) accomplishes this.
	// _, err = file.Seek(0, 0)
	// check(err)
	//
	// // The bufio package implements a buffered reader that may be useful both for its efficiency with many small reads and because of the additional reading methods it provides.
	// r4 := bufio.NewReader(f)
	// b4, err := r4.Peek(5)
	// check(err)
	// fmt.Printf("5 bytes: %s\n", string(b4))

}
