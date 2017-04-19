package favlist

import "testing"

func TestGetBucketInfo(t *testing.T) {
	binfo, err := getBucketID("11124261")
	if err != nil {
		t.Error(err)
	} else {
		for _, v := range binfo {
			t.Log(v)
		}
	}

}

func TestGetFavListInfo(t *testing.T) {
	binfo, _ := getBucketID("11124261")
	for _, v := range binfo {
		favlists, err := getFavList("11124261", v, nil)
		if err != nil {
			t.Error(err)
		} else {
			for _, vv := range favlists {
				t.Log(vv)
			}
		}
	}

}

func TestYaml(t *testing.T) {
	uid, err := readFavList()
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%v\n", uid)
	}
}

func TestGetAV(t *testing.T) {
	FavInfo, err := readFavList()
	if err != nil {
		t.Error(err)
	}
	err = FavInfo.Run()
	if err != nil {
		t.Error(err)
	}
}
