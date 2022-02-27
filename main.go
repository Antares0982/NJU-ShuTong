package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var EAI_SESS = os.Getenv("EAI_SESS")
var ROOM = strings.ReplaceAll(os.Getenv("ROOM"), " ", "")
var TGID = os.Getenv("TGID")
var TGBOTID = os.Getenv("TGBOTID")

func sendMessage(content string) (string, error) {
	url := "https://api.telegram.org/bot" + TGBOTID + "/sendMessage?chat_id=" + TGID + "&text=" + content
	request, err := http.Get(url)
	if err != nil {
		return "", err
	}

	bodyByte, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyByte)
	return body, nil
}

func doGet(url string, headers map[string]string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bodyByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyByte)
	return body, nil
}

func getElectricBalance(areaId string, buildId string, roomId string) (float64, error) {
	url := fmt.Sprintf("https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=%s&build_id=%s&room_id=%s", areaId, buildId, roomId)
	body, err := doGet(url, map[string]string{
		"Cookie": "eai-sess=" + EAI_SESS,
	})
	if err != nil {
		return 0, err
	}
	reg := regexp.MustCompile(`dianyue\w*:\w*"([\d\.]+)"`)
	result := reg.FindStringSubmatch(body)

	if len(result) != 2 {
		return 0, errors.New("EAI_SESS 失效了")
	}
	return strconv.ParseFloat(result[1], 64)
}

func getElectricInfo() {
	ids := strings.SplitN(ROOM, ",", 3)

	if len(ids) != 3 {
		sendMessage("寝室未配置或配置错误，或未配置 EAI_SESS，不开启寝室电量监测")
		panic(errors.New("寝室未配置或配置错误"))
	}

	electric, err := getElectricBalance(ids[0], ids[1], ids[2])

	log.Println("taskElectricBalance", electric, err)

	var info string
	if err != nil {
		info, err = sendMessage("获取寝室电量出错了：" + err.Error())
	} else if electric <= 15 {
		info, err = sendMessage("寝室电量不足 15 度啦～")
	}

	if err == nil {
		log.Println(info)
	}
}

func main() {
	getElectricInfo()
}
