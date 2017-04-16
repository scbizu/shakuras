package favlist

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

//FavInfo  read things from yaml
type FavInfo struct {
	Hub struct {
		UID int `yaml:"uid"`
	} `yaml:"hub"`
}

//BucketInfo  read the buckets data
type BucketInfo struct {
	Status bool `json:"status"`
	Data   struct {
		List []struct {
			FavBox   int    `json:"fav_box"`
			Name     string `json:"name"`
			MaxCount int    `json:"max_count"`
			Count    int    `json:"count"`
			Videos   []struct {
				AID int    `json:"aid"`
				Pic string `json:"pic"`
			} `json:"videos"`
		} `json:"list"`
		Count int `json:"count"`
	} `json:"data"`
}

//FavList read all of the Favlist video data
type FavList struct {
	Status bool `json:"status"`
	Data   struct {
		Vlist []struct {
			AID int `json:"aid"`
		} `json:"vlist"`
	} `json:"data"`
}

//read uid from yaml
func readFavList() (*FavInfo, error) {

	yamldata, err := ioutil.ReadFile("favlist.yaml")
	if err != nil {
		return nil, err
	}
	favinfo := new(FavInfo)
	err = yaml.Unmarshal(yamldata, favinfo)
	if err != nil {
		return nil, err
	}
	return favinfo, nil
}

//return rsc address
func analyseFavList(AID string) error {
	resp, err := http.Get("https://api.prprpr.me/dplayer/video/bilibili?aid=" + AID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	av, err := os.Create(AID + ".flv")
	if err != nil {
		return err
	}
	_, err = io.Copy(av, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getBucketID(UID string) ([]int, error) {

	rootURL, err := url.Parse("http://space.bilibili.com/ajax/fav/getBoxList?")
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("mid", UID)
	bucketURL := rootURL.String() + params.Encode()
	resp, err := http.Get(bucketURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bucketinfo := new(BucketInfo)
	err = json.Unmarshal(data, &bucketinfo)
	if err != nil {
		return nil, err
	}
	buckets := []int{}
	for _, v := range bucketinfo.Data.List {
		//its will save into the cache .
		buckets = append(buckets, v.FavBox)
	}
	return buckets, nil
}

func getFavList(UID string, BID int) ([]int, error) {
	rootURL, err := url.Parse("http://space.bilibili.com/ajax/fav/getList?")
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("mid", UID)
	params.Add("pagesize", "30")
	SBID := strconv.Itoa(BID)
	params.Add("fid", SBID)
	params.Add("tid", "0")
	params.Add("kw", "")
	params.Add("pid", "1")
	params.Add("order", "ftime")
	FavListURL := rootURL.String() + params.Encode()
	resp, err := http.Get(FavListURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	listData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	favlist := new(FavList)
	err = json.Unmarshal(listData, &favlist)
	if err != nil {
		return nil, err
	}
	avs := []int{}
	for _, v := range favlist.Data.Vlist {
		avs = append(avs, v.AID)
	}
	return avs, nil
}

//Run run the whole program
func (fav *FavInfo) Run() error {
	//get bucket
	var wg sync.WaitGroup

	avlists := []int{}
	binfo, err := getBucketID(strconv.Itoa(fav.Hub.UID))
	if err != nil {
		return err
	}
	for _, v := range binfo {
		favlists, er := getFavList(strconv.Itoa(fav.Hub.UID), v)
		if er != nil {
			return err
		}
		for _, avid := range favlists {
			avlists = append(avlists, avid)
		}
	}

	for _, av := range avlists {

		wg.Add(1)
		go func(vid int) {
			println(vid)
			defer wg.Done()
			err = analyseFavList(strconv.Itoa(vid))
			if err != nil {
				log.Fatalln(err.Error())
			}
		}(av)
	}
	wg.Wait()
	return nil
}
