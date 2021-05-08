package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eric11jhou/comdi/xls"
	"github.com/gin-gonic/gin"
)

type Option struct {
	Time string `json:"Time"`
	Put  One    `json:"Put"`
	Call One    `json:"Call"`
}

type One struct {
	Total []interface{}   `json:"Total"`
	Rank  [][]interface{} `json:"Rank"`
}

var CFTC_FILENAME = "com_disagg_xls_2021.zip"
var CFTC_XLSNAME = "c_year.xls"

type CFTD struct {
	Time  string      `time`
	Datas []*CFTCData `json:"datas"`
}

type CFTCData struct {
	Date                    string  `json:"date"`
	MMoneyPositionLongAll   int     `json:"m_money_long"`
	NonReptPositionLongAll  int     `json:"non_rept_long"`
	MMoneyPositionShortAll  int     `json:"m_money_short"`
	NonReptPositionShortAll int     `json:"non_rept_short"`
	Diff                    int     `json:"diff"`
	ChangePerWeek           int     `json:"change_per_week"`
	Percent                 float64 `json:"percent"`
}

func main() {
	go cron1d()
	r := gin.Default()
	r.GET("/option", func(c *gin.Context) {
		jsonFile, err := os.Open("data.json")
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var op Option
		err = json.Unmarshal(byteValue, &op)
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		c.JSON(200, op)
	})
	r.GET("/cftc", func(c *gin.Context) {
		jsonFile, err := os.Open("cftc.json")
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var cftc CFTD
		err = json.Unmarshal(byteValue, &cftc)
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		c.JSON(200, cftc)
	})
	r.Run(":80")
}

func cron1d() {
	fetchCFTC()
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))
		<-t.C
		go fetchCFTC()
	}
}

func fetchCFTC() {
	err := downloadCFTC()
	if err != nil {
		fmt.Println(err)
	}
	err = unzip(CFTC_FILENAME, "./")
	if err != nil {
		fmt.Println(err)
	}
	if xlFile, err := xls.Open(CFTC_XLSNAME, "utf-8"); err == nil {
		if sheet1 := xlFile.GetSheet(0); sheet1 != nil {
			row1 := sheet1.Row(0)
			DateIndex := 0
			MMoneyPositionLongAllIndex := 0
			NonReptPositionLongAllIndex := 0
			MMoneyPositionShortAllIndex := 0
			NonReptPositionShortAllIndex := 0
			for i := 0; i <= row1.LastCol(); i++ {
				switch row1.Col(i) {
				case "Report_Date_as_MM_DD_YYYY":
					DateIndex = i
				case "M_Money_Positions_Long_ALL":
					MMoneyPositionLongAllIndex = i
				case "NonRept_Positions_Long_All":
					NonReptPositionLongAllIndex = i
				case "M_Money_Positions_Short_ALL":
					MMoneyPositionShortAllIndex = i
				case "NonRept_Positions_Short_All":
					NonReptPositionShortAllIndex = i
				}
				if DateIndex != 0 &&
					MMoneyPositionLongAllIndex != 0 &&
					NonReptPositionLongAllIndex != 0 &&
					MMoneyPositionShortAllIndex != 0 &&
					NonReptPositionShortAllIndex != 0 {
					break
				}
			}
			CFTCDatas := make([]*CFTCData, 0)
			for i := 1; i <= int(sheet1.MaxRow); i++ {
				row := sheet1.Row(i)
				if row.Col(0) == "GOLD - COMMODITY EXCHANGE INC." {
					iMMoneyPositionLongAll, _ := strconv.Atoi(row.Col(MMoneyPositionLongAllIndex))
					iNonReptPositionLongAll, _ := strconv.Atoi(row.Col(NonReptPositionLongAllIndex))
					iMMoneyPositionShortAll, _ := strconv.Atoi(row.Col(MMoneyPositionShortAllIndex))
					iNonReptPositionShortAll, _ := strconv.Atoi(row.Col(NonReptPositionShortAllIndex))
					CFTCDatas = append([]*CFTCData{
						&CFTCData{
							Date:                    row.Col(DateIndex),
							MMoneyPositionLongAll:   iMMoneyPositionLongAll,
							NonReptPositionLongAll:  iNonReptPositionLongAll,
							MMoneyPositionShortAll:  iMMoneyPositionShortAll,
							NonReptPositionShortAll: iNonReptPositionShortAll,
							Diff:                    iMMoneyPositionLongAll + iNonReptPositionLongAll - iMMoneyPositionShortAll - iNonReptPositionShortAll,
						},
					}, CFTCDatas...)
				}
			}
			for i := range CFTCDatas {
				if i == 0 {
					continue
				}
				CFTCDatas[i].ChangePerWeek = CFTCDatas[i].Diff - CFTCDatas[i-1].Diff
				CFTCDatas[i].Percent = math.Abs(float64(CFTCDatas[i].ChangePerWeek) / float64(CFTCDatas[i-1].Diff))
			}
			CFTD := &CFTD{
				Time:  time.Now().Format("2006-01-02 15:04:05"),
				Datas: CFTCDatas,
			}
			datasJSON, _ := json.Marshal(CFTD)
			ioutil.WriteFile("cftc.json", datasJSON, 0644)
			fmt.Println("CFTC已更新: ", time.Now())
		}
	}
}

func downloadCFTC() error {
	specUrl := fmt.Sprintf("https://www.cftc.gov/files/dea/history/%s", CFTC_FILENAME)
	resp, err := http.Get(specUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(CFTC_FILENAME)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
