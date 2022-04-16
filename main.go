package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zc2638/ddshop/pkg/notice"
)

const TimeFormat = "2006/01/02 15:04:05"

var (
	longitude string
	latitude  string
	name      string
	bark      string
	sleep     uint64
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
		FullTimestamp:          true,
		TimestampFormat:        TimeFormat,
	})
}

func main() {
	command := getCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func getCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "vanguard-store",
		Short:        "华润万家门店列表程序",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if latitude == "" || longitude == "" {
				return errors.New("请输入经度和纬度.")
			}

			barKey := &notice.BarkConfig{Key: bark}
			barkIns := notice.NewBark(barKey)
			// music := notice.NewMusic(asserts.NoticeMP3, 180)
			// noticeIns := notice.New(notice.NewLog(), bark)
			// noticeIns.Notice("aa", "bb")

			if sleep <= 0 {
				sleep = 60
			}
			ticker := time.NewTicker(time.Second * time.Duration(sleep))

			for {
				Stores, err := getStores(longitude, latitude)
				if err != nil {
					logrus.Warningf("Get stores failure: %v", err)
				} else {
					if name != "" {
						filterFun := func(store StoreItem) bool {
							return strings.Contains(store.StoreName, name) || strings.Contains(store.StoresAddress, name)
						}
						var filterStores []StoreItem
						for _, store := range Stores {
							if filterFun(store) {
								filterStores = append(filterStores, store)
							}
						}
						if len(filterStores) > 0 {
							if str, err := json.Marshal(filterStores); err != nil {
								logrus.Warningf("Marshal filter stores failure: %v", err)
							} else {
								logrus.Infof("筛选结果:%s", str)
								if bark != "" {
									barkIns.Send("找到门店", string(str))
								}
							}
						} else {
							logrus.Warning("筛选结果为空")
						}
					} else {
						if str, err := json.Marshal(Stores); err != nil {
							logrus.Warningf("Marshal stores failure: %v", err)
						} else {
							logrus.Info(string(str))
						}
					}
				}

				<-ticker.C
			}
		},
	}

	cmd.Flags().StringVarP(&longitude, "long", "l", "", "经度")
	cmd.Flags().StringVarP(&latitude, "lat", "t", "", "纬度")
	cmd.Flags().StringVarP(&name, "name", "n", "", "过滤门店名称")
	cmd.Flags().StringVarP(&bark, "bark", "b", "", "Bark Key")
	cmd.Flags().Uint64VarP(&sleep, "sleep", "s", 60, "Sleep seconds")

	return cmd
}

func getStores(long, lat string) ([]StoreItem, error) {
	u, err := url.Parse("https://app.crv.com.cn/app_api/v1/dc-app-api/mobile/api/store/selectByAddress")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	paramVal := fmt.Sprintf(`{"longitude": "%s","latitude": "%s"}`, long, lat)
	params.Add("param", paramVal)
	u.RawQuery = params.Encode()
	urlPath := u.String()

	response, err := http.Get(urlPath)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.StateCode != 0 {
		return nil, errors.New("Http api request errors.")
	}

	return result.Data.Stores, nil
}

type Result struct {
	StateCode int `json:"code"`
	Data      struct {
		Stores []StoreItem `json:"stores"`
	} `json:"data"`
}

type StoreItem struct {
	AreaCode       string `json:"areaCode"`
	Buid           int    `json:"buid"`
	CityId         int    `json:"cityId"`
	CityName       string `json:"cityName"`
	HqId           string `json:"hq_id"`
	RetailFormatId string `json:"retailFormatId"`
	StoreId        int    `json:"storeId"`
	StoreName      string `json:"storeName"`
	StoresAddress  string `json:"storesAddress"`
}
