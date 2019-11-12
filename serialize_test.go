package cronrange

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCronRange_String(t *testing.T) {
	tests := []struct {
		name string
		cr   *CronRange
		want string
	}{
		{"Nil struct", crNil, "<nil>"},
		{"Empty struct", crEmpty, emptyString},
		{"Use string() instead of sprintf", crEvery1Min, "DR=1; * * * * *"},
		{"Use instance instead of pointer", crEvery1Min, "DR=1; * * * * *"},
		{"1min duration without time zone", crEvery1Min, "DR=1; * * * * *"},
		{"5min duration without time zone", crEvery5Min, "DR=5; */5 * * * *"},
		{"10min duration with local time zone", crEvery10MinLocal, "DR=10; */10 * * * *"},
		{"10min duration with time zone", crEvery10MinBangkok, "DR=10; TZ=Asia/Bangkok; */10 * * * *"},
		{"Every xmas morning in new york city", crEveryXmasMorningNYC, "DR=240; TZ=America/New_York; 0 8 25 12 *"},
		{"Every new year's day in bangkok", crEveryNewYearsDayBangkok, "DR=1440; TZ=Asia/Bangkok; 0 0 1 1 *"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := tt.cr
			var got string
			if strings.Contains(tt.name, "string()") {
				got = cr.String()
			} else if strings.Contains(tt.name, "instance") {
				got = fmt.Sprint(*cr)
			} else {
				got = fmt.Sprint(cr)
			}
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkCronRange_String(b *testing.B) {
	cr, _ := New(exprEveryMin, timeZoneBangkok, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cr.String()
	}
}

var deserializeTestCases = []struct {
	name    string
	inputS  string
	wantS   string
	wantErr bool
}{
	{"Empty string", emptyString, emptyString, true},
	{"Invalid expression", "hello", emptyString, true},
	{"Missing duration", "; * * * * *", emptyString, true},
	{"Invalid duration=0", "DR=0;* * * * *", emptyString, true},
	{"Invalid duration=-5", "DR=-5;* * * * *", emptyString, true},
	{"Invalid with Mars time zone", "DR=5;TZ=Mars;* * * * *", emptyString, true},
	{"Invalid with unknown part", "DR=10; TZ=Pacific/Honolulu; SET=1; * * * * *", emptyString, true},
	{"Invalid with lower case", "dr=5;* * * * *", emptyString, true},
	{"Invalid with wrong order", "* * * * *; DR=5;", emptyString, true},
	{"Normal without timezone", "DR=5;* * * * *", "DR=5; * * * * *", false},
	{"Normal with extra whitespaces", "  DR=6 ;  * * * * *  ", "DR=6; * * * * *", false},
	{"Normal with empty parts", ";  DR=7;;; ;; ;; ;* * * * *  ", "DR=7; * * * * *", false},
	{"Normal with local time zone", "DR=8;TZ=Local;* * * * *", "DR=8; * * * * *", false},
	{"Normal with UTC time zone", "DR=9;TZ=Etc/UTC;* * * * *", "DR=9; TZ=Etc/UTC; * * * * *", false},
	{"Normal with Honolulu time zone", "DR=10;TZ=Pacific/Honolulu;* * * * *", "DR=10; TZ=Pacific/Honolulu; * * * * *", false},
	{"Normal with Honolulu time zone in different order", "TZ=Pacific/Honolulu; DR=10; * * * * *", "DR=10; TZ=Pacific/Honolulu; * * * * *", false},
	{"Normal with complicated expression", "DR=5258765;   TZ=Pacific/Honolulu;   4,8,22,27,33,38,47,50 3,11,14-16,19,21,22 */10 1,3,5,6,9-11 1-5", "DR=5258765; TZ=Pacific/Honolulu; 4,8,22,27,33,38,47,50 3,11,14-16,19,21,22 */10 1,3,5,6,9-11 1-5", false},
}

func TestParseString(t *testing.T) {
	for _, tt := range deserializeTestCases {
		t.Run(tt.name, func(t *testing.T) {
			gotCr, err := ParseString(tt.inputS)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error: %v, wantErr: %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (gotCr == nil || gotCr.schedule == nil || gotCr.duration == 0) {
				t.Errorf("ParseString() incomplete gotCr: %v", gotCr)
				return
			}
			if !tt.wantErr && gotCr.String() != tt.wantS {
				t.Errorf("ParseString() gotCr: %s, want: %s", gotCr.String(), tt.wantS)
			}
		})
	}
}

func BenchmarkParseString(b *testing.B) {
	rs := "DR=10;TZ=Pacific/Honolulu;;* * * * *"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseString(rs)
	}
}

func TestCronRange_MarshalJSON(t *testing.T) {
	tempStruct := tempTestStruct{
		nil,
		"Test",
		1111,
	}
	tests := []struct {
		name  string
		cr    *CronRange
		wantJ string
	}{
		{"Nil struct", crNil, `{"CR":null,"Name":"Test","Value":1111}`},
		{"Empty struct", crEmpty, `{"CR":null,"Name":"Test","Value":1111}`},
		{"5min duration without time zone", crEvery5Min, `{"CR":"DR=5; */5 * * * *","Name":"Test","Value":1111}`},
		{"10min duration with local time zone", crEvery10MinLocal, `{"CR":"DR=10; */10 * * * *","Name":"Test","Value":1111}`},
		{"10min duration with time zone", crEvery10MinBangkok, `{"CR":"DR=10; TZ=Asia/Bangkok; */10 * * * *","Name":"Test","Value":1111}`},
		{"Every xmas morning in new york city", crEveryXmasMorningNYC, `{"CR":"DR=240; TZ=America/New_York; 0 8 25 12 *","Name":"Test","Value":1111}`},
		{"Every new year's day in bangkok", crEveryNewYearsDayBangkok, `{"CR":"DR=1440; TZ=Asia/Bangkok; 0 0 1 1 *","Name":"Test","Value":1111}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempStruct.CR = tt.cr
			got, err := json.Marshal(tempStruct)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			gotJ := string(got)
			if gotJ != tt.wantJ {
				t.Errorf("MarshalJSON() got = %v, want %v", gotJ, tt.wantJ)
			}
		})
	}
}

func BenchmarkCronRange_MarshalJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = crEvery10MinBangkok.MarshalJSON()
	}
}

func TestCronRange_UnmarshalJSON(t *testing.T) {
	jsonPrefix, jsonSuffix := `{"CR":"`, `","Name":"Demo","Value":2222}`
	for _, tt := range deserializeTestCases {
		t.Run(tt.name, func(t *testing.T) {
			jsonFull := jsonPrefix + tt.inputS + jsonSuffix
			var gotS tempTestStruct
			err := json.Unmarshal([]byte(jsonFull), &gotS)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error: %v, wantErr: %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (gotS.CR == nil || gotS.CR.schedule == nil || gotS.CR.duration == 0) {
				t.Errorf("UnmarshalJSON() incomplete gotCr: %v", gotS.CR)
				return
			}
			if !tt.wantErr && gotS.CR.String() != tt.wantS {
				t.Errorf("UnmarshalJSON() gotCr: %s, want: %s", gotS.CR.String(), tt.wantS)
				return
			}

			jsonBrokens := []string{
				jsonPrefix[0:len(jsonPrefix)-1] + tt.inputS + jsonSuffix[1:len(jsonSuffix)-1],
				jsonPrefix[0:len(jsonPrefix)-1] + tt.inputS + jsonSuffix,
				jsonPrefix + tt.inputS + jsonSuffix[1:len(jsonSuffix)-1],
				jsonSuffix + jsonPrefix,
				jsonPrefix + tt.inputS,
				tt.inputS + jsonSuffix,
				tt.inputS + jsonPrefix,
				jsonSuffix + tt.inputS,
				jsonSuffix + tt.inputS + jsonPrefix,
				tt.inputS + jsonSuffix + jsonPrefix,
				jsonSuffix + jsonPrefix + tt.inputS,
				tt.inputS + jsonPrefix + jsonSuffix,
				jsonPrefix + jsonSuffix + tt.inputS,
			}
			for _, jsonBroken := range jsonBrokens {
				if err = json.Unmarshal([]byte(jsonBroken), &gotS); err == nil {
					t.Errorf("UnmarshalJSON() missing error for broken json: %s", jsonBroken)
					return
				}
			}
		})
	}
}

func BenchmarkCronRange_UnmarshalJSON(b *testing.B) {
	jsonFull := []byte(`{"CR":"DR=10;TZ=Pacific/Honolulu;* * * * *","Name":"Demo","Value":2222}`)
	var gotS tempTestStruct
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(jsonFull, &gotS)
	}
}

func TestTimeRange_String(t *testing.T) {
	type fields struct {
		Start time.Time
		End   time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"From zero to zero", fields{zeroTime, zeroTime}, "[0001-01-01T00:00:00Z,0001-01-01T00:00:00Z]"},
		{"First day of 2020 in utc", fields{firstSec2020Utc, firstSec2020Utc.AddDate(0, 0, 1)}, "[2020-01-01T00:00:00Z,2020-01-02T00:00:00Z]"},
		{"First month of 2019 in bangkok", fields{firstSec2019Bangkok, firstSec2019Bangkok.AddDate(0, 1, 0)}, "[2019-01-01T00:00:00+07:00,2019-02-01T00:00:00+07:00]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TimeRange{
				Start: tt.fields.Start,
				End:   tt.fields.End,
			}
			if got := tr.String(); got != tt.want {
				t.Errorf("String() = %v, want = %v", got, tt.want)
			}
		})
	}
}

func BenchmarkTimeRange_String(b *testing.B) {
	tr := TimeRange{firstSec2019Bangkok, firstSec2019Bangkok.AddDate(0, 1, 0)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.String()
	}
}