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
		favlists, err := getFavList("11124261", v)
		if err != nil {
			t.Error(err)
		} else {
			for _, vv := range favlists {
				t.Log(vv)
			}
		}
	}

}

func TestGetAV(t *testing.T) {
	err := analyseFavList("8059894")
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Downloaded")
	}
}
