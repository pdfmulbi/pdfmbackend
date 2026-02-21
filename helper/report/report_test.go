package report

import (
	"fmt"
	"os"
	"testing"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
)

func TestGenerateReport(t *testing.T) {
	t.Skip("Skipping DB test")
	mongoinfo := atdb.DBInfo{
		DBString: os.Getenv("MONGODOMYID"),
		DBName:   "domyid",
	}
	Mongoconn, _ := atdb.MongoConnect(mongoinfo)
	config.WAAPIToken = "v4.public."
	fmt.Println(mongoinfo.DBString)
	err := RekapMeetingKemarin(Mongoconn)
	fmt.Println(err)
}

/* func TestGenerateReportLayanan(t *testing.T) {
	gid := "6281313112053-1492882006"
	results := GetDataLaporanMasukHariini(Mongoconn, gid) //GetDataLaporanMasukHarian
	print(results)

}

func TestGenerateReportLay(t *testing.T) {
	//gid := "6281313112053-1492882006"
	results := GetDataLaporanMasukHarian(Mongoconn) //GetDataLaporanMasukHarian
	print(results)

} */
