// utils project utils.go
package utils

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/ulikunitz/xz"

	"paleotronic.com/fmt"
)

type Foo int

func init() {
	SeedRandom()
}

func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func GetSystemUser() string {

	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME")
	}
	return os.Getenv("USER")

}

func GetSystemHost() string {

	if runtime.GOOS == "windows" {
		return os.Getenv("COMPUTERNAME")
	}
	return os.Getenv("HOSTNAME")

}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetUniqueName() string {
	return RandStringRunes(16)
}

func InList(s string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(s) == strings.ToLower(v) {
			return true
		}
	}
	return false
}

func SplitLines(bb []byte) []string {
	data := string(bb)
	var sl []string
	var chunk string
	for _, r := range data {
		if r == 13 || r == 10 {
			if len(chunk) > 0 {
				sl = append(sl, chunk)
				chunk = ""
			}
		} else {
			chunk = chunk + string(r)
		}
	}
	if len(chunk) > 0 {
		sl = append(sl, chunk)
		chunk = ""
	}

	return sl
}

// Length in unicode code points
func Len(str string) int {
	return utf8.RuneCount([]byte(str))
}

// Copy based on unicode string
func Copy(str string, start int, count int) string {
	sindex := start - 1
	eindex := sindex + count

	if sindex > len(str) {
		return ""
	}

	if eindex < sindex {
		return ""
	}

	if eindex > len(str) {
		eindex = len(str)
	}

	out := ""
	rpos := 0
	for _, r := range str {
		if rpos >= sindex && rpos < eindex {
			out += string(r)
		}
		rpos++
	}

	return out
}

func Delete(str string, start int, count int) string {
	sindex := start - 1
	eindex := sindex + count

	if sindex > len(str) {
		return str
	}

	if eindex < sindex {
		return str
	}

	if eindex > len(str) {
		eindex = len(str)
	}

	out := ""
	rpos := 0
	for _, r := range str {
		if rpos < sindex || rpos >= eindex {
			out += string(r)
		}
		rpos++
	}

	return out
}

// Return
func Pos(substr, str string) int {
	return strings.Index(str, substr) + 1
}

// Return the position of a rune within the given string, else 0
func PosRune(char rune, str string) int {
	return strings.IndexRune(str, char) + 1
}

// Strip bit 7 from character string.  Used to convert strings in integer basic to ascii form.
func FlattenASCII(in string) string {
	var out string
	for _, ch := range in {
		ch = ch & 127
		out = out + string(ch)
	}
	return out
}

// Alias for FlattenASCII
func Flatten7Bit(in string) string {
	return FlattenASCII(in)
}

// Return the numeric part of a string
func NumberPart(in string) string {
	var out string
	for _, ch := range in {
		if (ch >= '0') && (ch <= '9') {
			out = out + string(ch)
			continue
		}
		if (ch == '-') || (ch == 'e') || (ch == 'E') || (ch == '+') || (ch == '.') {
			out = out + string(ch)
			continue
		}
		if ch == ' ' {
			continue
		}
		break
	}
	return out
}

// Seed the random number generator based on current nano seconds
func SeedRandom() {
	rand.Seed(time.Now().UnixNano())
}

func PSeed(v int64) {
	rand.Seed(v)
}

// Return a random number between 0 and 1
func Random() float64 {
	return rand.Float64()
}

func StrToIntStr(v string) string {
	f, _ := strconv.ParseFloat(NumberPart(FlattenASCII(v)), 32)

	return strconv.FormatFloat(math.Floor(f), 'f', 0, 32)
}

func StrToFloatStr(v string) string {
	f, _ := strconv.ParseFloat(NumberPart(FlattenASCII(v)), 64)
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func StrToFloat(v string) float32 {
	f, _ := strconv.ParseFloat(NumberPart(FlattenASCII(v)), 64)
	return float32(f)
}

func StrToFloat64(v string) float64 {
	f, _ := strconv.ParseFloat(NumberPart(FlattenASCII(v)), 64)
	return f
}

func IntToStr(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func FloatToStr(i float64) string {
	return strconv.FormatFloat(i, 'f', -1, 64)
}

func FormatFloat(f string, i float64) string {
	return FloatToStr(i)
}

func ReadTextFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func WriteTextFile(path string, content []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	for _, l := range content {
		file.WriteString(l + "\r\n")
	}
	return file.Close()
}

func WriteBinaryFile(path string, content []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Write(content)
	return file.Close()
}

func AppendTextFile(path string, content []string) error {
	ss, err := ReadTextFile(path)
	if err != nil {
		return err
	}
	ss = append(ss, content...)
	return WriteTextFile(path, ss)
}

func StrToInt(s string) int {
	ii, _ := strconv.ParseInt(s, 0, 32)
	return int(ii)
}

func IntSliceEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func XZBytes(data []byte) []byte {
	var b bytes.Buffer
	w, _ := xz.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func UnXZBytes(data []byte) []byte {
	b := bytes.NewBuffer(data)
	out := bytes.NewBuffer([]byte(nil))
	r, _ := xz.NewReader(b)
	io.Copy(out, r)
	return out.Bytes()
}

func GZIPBytes(data []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func UnGZIPBytes(data []byte) []byte {
	b := bytes.NewBuffer(data)
	out := bytes.NewBuffer([]byte(nil))
	r, _ := gzip.NewReader(b)
	io.Copy(out, r)
	r.Close()
	return out.Bytes()
}

func FloatToStrApple(n float64) string {
	// to convert a float number to a string

	if n > 100000000 {
		return strconv.FormatFloat(n, 'g', 9, 64)
	}

	if n < 0.01 {
		return strconv.FormatFloat(n, 'e', -1, 64)
	}

	return strconv.FormatFloat(n, 'f', -1, 64)
}

func StrToFloatStrApple(s string) string {
	n, _ := strconv.ParseFloat(s, 64)

	if n == 0 {
		return "0"
	}

	if math.Abs(n) > 100000000 {
		return strconv.FormatFloat(n, 'g', 9, 64)
	}

	if math.Abs(n) < 0.01 {
		return strconv.FormatFloat(n, 'e', -1, 64)
	}

	//return strconv.FormatFloat(n, 'f', -1, 64)
	return SignificantDigits(n, 9)
}

func StrToFloatStrAppleLogo(s string) string {
	n, _ := strconv.ParseFloat(s, 64)

	if math.Abs(n) < 0.0000005 {
		return "0"
	}

	str := strconv.FormatFloat(n, 'f', 6, 64)
	str = strings.TrimRight(str, "0")

	if strings.HasSuffix(str, ".") {
		str = strings.TrimRight(str, ".")
	}

	return str
}

func SignificantDigits(f float64, digits int) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)

	dc := 0
	dcap := 0
	dcbp := 0
	afterPoint := false
	nzBP := false
	ns := ""

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			if dc < digits {
				if afterPoint {
					dcap++
				} else {
					if ch > '0' {
						nzBP = true
					}
					if nzBP {
						dcbp++
					}
				}
				ns += string(ch)
			} else if !afterPoint {
				dcbp++
				ns += "0"
			}
			if afterPoint || nzBP {
				dc++
			}
		} else {
			if ch == '.' {
				afterPoint = true
			}
			ns += string(ch)
		}
	}

	fmtstr := fmt.Sprintf("%%%d.%df", dcbp, dcap)

	ns = fmt.Sprintf(fmtstr, f)

	if strings.HasPrefix(ns, "-0.") {
		ns = "-" + ns[2:]
	} else if strings.HasPrefix(ns, "0.") {
		ns = ns[1:]
	}
	for len(ns) > 1 && strings.HasSuffix(ns, "0") && strings.Contains(ns, ".") {
		ns = ns[0 : len(ns)-1]
	}
	if strings.HasSuffix(ns, ".") {
		ns = ns[0 : len(ns)-1]
	}

	return ns
}

var reHexNumber = regexp.MustCompile("^(0x|[$])([0-9a-fA-F]+)$")
var reDecNumber = regexp.MustCompile("^([0-9]+)$")

func SuperStrToInt(in string) (int, error) {

	if reHexNumber.MatchString(in) {
		m := reHexNumber.FindAllStringSubmatch(in, -1)
		//fmt.Println("0x", m[0][2])
		return StrToInt("0x" + m[0][2]), nil
	}

	if reDecNumber.MatchString(in) {
		m := reDecNumber.FindAllStringSubmatch(in, -1)
		//fmt.Println(m[0][1])
		return StrToInt(m[0][1]), nil
	}

	return -1, errors.New("NaN: " + in)
}

func Overflow(s string, mark string, maxlen int) string {
	if len(s) > maxlen {
		s = s[:maxlen-len(mark)]
		s += mark
	}
	return s
}

func TurtleCoordinatesToGL(glWidth, glHeight, width, height float64, lpos, tpos mgl64.Vec3) mgl64.Vec3 {

	unitWidth := glWidth / width
	unitHeight := glHeight / height
	unitDepth := unitHeight

	hx, hy := (unitWidth * width / 2), (unitHeight * height / 2)

	tpos[0] = tpos[0]*unitWidth + hx + lpos[0]
	tpos[1] = tpos[1]*unitHeight + hy + lpos[1]
	tpos[2] = tpos[2]*unitDepth + lpos[2]

	return tpos
}
