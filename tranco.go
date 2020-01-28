package tranco

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

type Tranco struct {
	Should_cache bool
	Cache_dir    string
}
type TrancoList struct {
	Date         string
	List_id      string
	List_page    string
	Domains_list []Domain
	Domains_map  map[string]int
}

type Domain struct {
	Rank int
	Name string
}

func (tl TrancoList) Top(num int) []Domain {
	return tl.Domains_list[:num]

}
func (tl TrancoList) Rank(domainname string) int {
	rank, ok := tl.Domains_map[domainname]
	if ok {
		return rank
	} else {
		return 0
	}
}
func getListIdForDate(date string) string {

	url := "https://tranco-list.eu/daily_list_id?date=" + date
	resp, err := http.Get(url)
	checkError("Error getting listId. ", err)
	defer resp.Body.Close()

	body, readerr := ioutil.ReadAll(resp.Body)
	checkError("Error with reading http reponse body", readerr)
	return string(body)

}
func downloadZipFile(filepath string, list_id string) error {
	download_url := "https://tranco-list.eu/download_daily/" + list_id
	resp, err := http.Get(download_url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err

}
func (t Tranco) List(date string) TrancoList {
	var tl TrancoList
	var list_id string
	var filepath string
	if date == "latest" {
		yesterday := time.Now().AddDate(0, 0, -1)
		date = yesterday.Format("2006-01-02")
	}
	list_id = getListIdForDate(date)

	if t.Should_cache {
		if t.Cache_dir == "" {
			wd, err := os.Getwd()
			checkError("Error with getting working dir!", err)
			cachepath := path.Join(wd, ".tranco")
			t.Cache_dir = cachepath
		}
		if _, err := os.Stat(t.Cache_dir); os.IsNotExist(err) {
			err = os.Mkdir(t.Cache_dir, 0744)
			checkError("Error with creating cache folder", err)
		}
		filename := list_id + ".zip"
		filepath = path.Join(t.Cache_dir, filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {

			err := downloadZipFile(filepath, list_id)
			checkError("Cannot download file from tranco-list", err)
		}
	} else {
		filepath = list_id + ".zip"
		err := downloadZipFile(filepath, list_id)
		checkError("Cannot download file from tranco-list", err)
	}
	err := unzipfile(filepath, "/tmp/tranco")
	checkError("Error with unzipping file!", err)

	// read csv file
	csvFile, err := os.Open("/tmp/tranco/top-1m.csv")
	checkError("Error with opening top-1m.csv file", err)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	tl.Domains_map = make(map[string]int)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		checkError("Error with reading csv file!", err)
		rank, err := strconv.Atoi(line[0])
		checkError("Error with converting rank to int!", err)

		tl.Domains_list = append(tl.Domains_list, Domain{Rank: rank, Name: line[1]})
		tl.Domains_map[line[1]] = rank
	}
	tl.Date = date
	tl.List_id = list_id
	tl.List_page = "https://tranco-list.eu/list/" + list_id + "/1000000"
	return tl
}

func unzipfile(filepath string, target string) error {
	reader, err := zip.OpenReader(filepath)
	checkError("Error with reading zip file", err)
	if err = os.MkdirAll(target, 0744); err != nil {
		return err
	}
	for _, file := range reader.File {
		zippedfile := path.Join(target, file.Name)
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		unzippedfile, err := os.OpenFile(zippedfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer unzippedfile.Close()
		if _, err := io.Copy(unzippedfile, fileReader); err != nil {
			return err
		}

	}
	return nil

}
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
