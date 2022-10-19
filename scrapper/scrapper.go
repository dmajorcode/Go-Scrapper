package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct{
	id string
	name string
	summary string
}

// Scrape Saramin
func Scrape(){
	var baseURL string = "https://www.saramin.co.kr/zf_user/jobs/public/list/"
	var jobs []extractedJob
	c:=make(chan []extractedJob)
	totalPages:=getPages(baseURL)
	
	for i :=0;i<totalPages;i++{
		go getPage(i, baseURL, c)	
	}

	for i :=0; i<totalPages;i++{
		extractedJobs:=<-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainC chan <-[]extractedJob){
	var jobs[]extractedJob
	c:=make(chan extractedJob)
	pageURL := url + "page/" + strconv.Itoa(page)+"?sort=ud&listType=public&public_list_flag=y#searchTitle"
	fmt.Println("Requesting", pageURL)
	res, err :=http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err :=goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards :=doc.Find(".noti_list")
	
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
	})

	for i:=0; i<searchCards.Length();i++{
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob){
	id, _ := card.Attr("id")
	name :=CleanString(card.Find(".company_nm>a").Text())
	summary :=CleanString(card.Find(".job_tit>a").Text())
	c<-extractedJob{id:id, name:name, summary:summary,}
	
}

func getPages(url string) int{
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err :=goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages=s.Find("a").Length()
	})

	return pages
}

func checkErr(err error){
	if err!=nil{
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response){
	if res.StatusCode!=200{
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

// cleanString cleans a string
func CleanString(str string)string{
	return strings.Join(strings.Fields(strings.TrimSpace(str))," ")
}

func writeJobs(jobs []extractedJob){
	file, err:=os.Create("jobs.csv")
	checkErr(err)

	w:=csv.NewWriter(file)	
	defer w.Flush()

	headers:=[]string{"Name","Summary"}

	wErr:= w.Write(headers)
	checkErr(wErr)

	for _, job:=range jobs{
		jobSlice:=[]string {job.name, job.summary}
		jwErr :=w.Write(jobSlice)
		checkErr(jwErr)
	}
}