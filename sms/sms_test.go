package sms_test

import (
	"fmt"
	"testing"

	"github.com/suiguo/hwlib/sms"
)

type InterfaceTest struct {
	Name string
	Idx  int
}

func Test(t *testing.T) {
	// testArr := make([]InterfaceTest, 10)
	// for i := 0; i < 10; i++ {
	// 	testArr[i].Name = fmt.Sprintf("idname=%d", i)
	// }
	// tmp := map[stringt][]
	// phonenumbers.GetNddPrefixForRegion()
	// number := phonenumbers.GetRegionCodeForCountryCode(86)
	cli, err := sms.GetSmsClient(
		sms.Twilio,
		sms.WithAccount("ACdd2ab18445c4784ca24fbda9b7bc15a5"),
		sms.WithMsgId("MG196caf38adff19a1e90efe36706bcf4c"),
		sms.WithToken("e32a44218d52ff024812c62e1728c371"))
	fmt.Println(err)
	err = cli.SendSms("+86 17612199113", "13241")
	fmt.Println(err)
}
