package selenium
import (
	"sourcegraph.com/sourcegraph/go-selenium"
	"testing"
	"github.com/ruslanBik4/httpgo/models/logs"
)

var caps selenium.Capabilities
var executorURL = "http://vps-20777.vps-default-host.net"

// An example test using the WebDriverT and WebElementT interfaces. If you use the non-*T
// interfaces, you must perform error checking that is tangential to what you are testing,
// and you have to destructure results from method calls.
func TestWithT(t *testing.T) {
	wd, err := selenium.NewRemote(caps, executorURL)
	if (err != nil) {
		t.Errorf("err= %s", err)
	}
	// Call .T(t) to obtain a WebDriverT from a WebDriver (or to obtain a WebElementT from
	// a WebElement).
	wdt := wd.T(t)
	logs.DebugLog("wdt=",wdt)
	// Calls `t.Fatalf("Get: %s", err)` upon failure.
	wdt.Get("http://vps-20777.vps-default-host.net/extranet/objects/")

	//wdt.manage().window().maximize();
	//.findElement(By.Id("searchField")).sendKeys("Голден Резорт");
	// Calls `t.Fatalf("FindElement(by=%q, value=%q): %s", by, value, err)` upon failure.
	elem := wdt.FindElement(selenium.ById, "searchField")
	//t.Logf("want elem text %q, got %q", "bar","ddd")
	t.Logf("want elem text %s, got %q", elem)
	// Calls `t.Fatalf("Text: %s", err)` if the `.Text()` call fails.
	//if elem.GetAttribute("placeholder") != "Что искать?" {
	//	t.Fatalf("want elem text %q, got %q", "bar", elem.Text())
	//}
}
