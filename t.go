package testdriver

import (
	"github.com/daaku/go.testdriver/testing"
	"github.com/tebeka/selenium"
	"net/url"
)

type T struct {
	Driver           selenium.WebDriver
	T                *testing.T
	mainWindowHandle string
}

type Element struct {
	El selenium.WebElement
	T  *testing.T
}

func (t *T) Fail() {
	t.T.Fail()
}

func (t *T) Failed() bool {
	return t.T.Failed()
}

func (t *T) FailNow() {
	t.T.FailNow()
}

func (t *T) Log(args ...interface{}) {
	t.T.Log(args...)
}

func (t *T) Logf(format string, args ...interface{}) {
	t.T.Logf(format, args...)
}

func (t *T) Error(args ...interface{}) {
	t.T.Error(args...)
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.T.Errorf(format, args...)
}

func (t *T) Fatal(args ...interface{}) {
	t.T.Fatal(args...)
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.T.Fatalf(format, args...)
}

func (t *T) Element(selector string) *Element {
	el, err := t.Driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		t.Fatalf("Failed to find element with selector %s with error %s",
			selector, err)
	}
	return &Element{El: el, T: t.T}
}

func (t *T) Get(url string) {
	err := t.Driver.Get(url)
	if err != nil {
		t.Fatalf("Failed to get URL %s with error %s", url, err)
	}
}

func (t *T) Title() string {
	title, err := t.Driver.Title()
	if err != nil {
		t.Fatalf("Failed to get page title with error %s", err)
	}
	return title
}

func (t *T) CurrentWindowHandle() string {
	currentWindowHandle, err := t.Driver.CurrentWindowHandle()
	if err != nil {
		t.Fatalf("Failed to get current window handle with error %s", err)
	}
	return currentWindowHandle
}

func (t *T) WindowHandles() []string {
	windowHandles, err := t.Driver.WindowHandles()
	if err != nil {
		t.Fatalf("Failed to get window handles with error %s", err)
	}
	return windowHandles
}

func (t *T) PageSource() string {
	pageSource, err := t.Driver.PageSource()
	if err != nil {
		t.Fatalf("Failed to get page source with error %s", err)
	}
	return pageSource
}

func (t *T) CurrentURL() *url.URL {
	urlString, err := t.Driver.CurrentURL()
	if err != nil {
		t.Fatalf("Failed to get current URL with error %s", err)
	}
	currentURL, err := url.Parse(urlString)
	if err != nil {
		t.Fatalf("Failed to parse current URL %s with error %s", urlString, err)
	}
	return currentURL
}

func (t *T) SwitchWindow(window string) {
	err := t.Driver.SwitchWindow(window)
	if err != nil {
		t.Fatalf("Failed to switch to window %s with error %s", window, err)
	}
}

func (t *T) SwitchFrame(frame string) {
	err := t.Driver.SwitchFrame(frame)
	if err != nil {
		t.Fatalf("Failed to switch to frame %s with error %s", frame, err)
	}
}

func (t *T) SwitchFrameBySource(prefix string) {
	out := t.ExecuteScript(
		`var iframes = document.getElementsByTagName('iframe');
		 for (var i=0; i<iframes.length; i++) {
		   var src = iframes[i].src;
		   src = src.substring(src.indexOf('/', 9));
		   if (src.indexOf(arguments[0]) === 0) {
		     if (iframes[i].name) {
		       return iframes[i].name;
		     }
		     return "";
		   }
		 }
		 return "";`,
		prefix)
	name, isString := out.(string)
	if !isString {
		t.Fatalf("Was expecting string name for frame but got %v of type %T",
			out, out)
	}
	if name == "" {
		t.Fatalf("Could not find frame with source prefix %s", prefix)
	}
	t.SwitchFrame(name)
}

func (t *T) ExecuteScriptAsync(script string, args ...interface{}) interface{} {
	out, err := t.Driver.ExecuteScriptAsync(script, args)
	if err != nil {
		t.Fatalf("Failed to execute async script with error %s:\n%s", err, script)
	}
	return out
}

func (t *T) ExecuteScript(script string, args ...interface{}) interface{} {
	out, err := t.Driver.ExecuteScript(script, args)
	if err != nil {
		t.Fatalf("Failed to execute script with error %s:\n%s", err, script)
	}
	return out
}

func (t *T) ExecuteScriptAsyncRaw(script string, args ...interface{}) []byte {
	out, err := t.Driver.ExecuteScriptAsyncRaw(script, args)
	if err != nil {
		t.Fatalf("Failed to execute async script with error %s:\n%s", err, script)
	}
	return out
}

func (t *T) ExecuteScriptRaw(script string, args ...interface{}) []byte {
	out, err := t.Driver.ExecuteScriptRaw(script, args)
	if err != nil {
		t.Fatalf("Failed to execute script with error %s:\n%s", err, script)
	}
	return out
}

func (t *T) DecodeElement(data []byte) *Element {
	element, err := t.Driver.DecodeElement(data)
	if err != nil {
		t.Fatalf("Failed to decode element: %v %T %s", data, data, err)
	}
	return &Element{T: t.T, El: element}
}

func (t *T) DecodeElements(data []byte) []*Element {
	elements, err := t.Driver.DecodeElements(data)
	if err != nil {
		t.Fatalf("Failed to decode elements: %v %T %s", data, data, err)
	}

	wrapped := make([]*Element, len(elements))
	for index, el := range elements {
		wrapped[index] = &Element{T: t.T, El: el}
	}
	return wrapped
}

func (t *T) SwitchPopupWindow() {
	if t.mainWindowHandle == "" {
		t.mainWindowHandle = t.CurrentWindowHandle()
	}

	windowHandles := t.WindowHandles()
	if len(windowHandles) != 2 {
		t.Fatalf("SwitchPopupWindow expects exactly two windows, found: %d",
			len(windowHandles))
	}
	for _, handle := range windowHandles {
		if handle != t.mainWindowHandle {
			t.SwitchWindow(handle)
			return
		}
	}
	t.Fatal("Failed to find other window.")
}

func (t *T) SwitchMainWindow() {
	if t.mainWindowHandle == "" {
		t.Fatal("Never left main window.")
	}
	t.SwitchWindow(t.mainWindowHandle)
}

func (e *Element) Text() string {
	text, err := e.El.Text()
	if err != nil {
		e.T.Fatalf("Failed to get text of element with error %s", err)
	}
	return text
}

func (e *Element) Click() {
	err := e.El.Click()
	if err != nil {
		e.T.Fatalf("Failed to click element with error %s", err)
	}
}

func (e *Element) SendKeys(keys string) {
	err := e.El.SendKeys(keys)
	if err != nil {
		e.T.Fatalf("Failed to send keys with error %s", err)
	}
}

func (e *Element) GetAttribute(name string) string {
	value, err := e.El.GetAttribute(name)
	if err != nil {
		e.T.Fatalf("Failed to get attribute %s with error %s", name, err)
	}
	return value
}

func (e *Element) IsDisplayed() bool {
	displayed, err := e.El.IsDisplayed()
	if err != nil {
		e.T.Fatalf("Failed to check if element is displayed with error %s", err)
	}
	return displayed
}

func (e *Element) Element(selector string) *Element {
	el, err := e.El.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		e.T.Fatalf("Failed to find element with selector %s with error %s",
			selector, err)
	}
	return &Element{El: el, T: e.T}
}

func (e *Element) Submit() {
	err := e.El.Submit()
	if err != nil {
		e.T.Fatalf("Failed to submit element with error %s", err)
	}
}

func (e *Element) Clear() {
	err := e.El.Clear()
	if err != nil {
		e.T.Fatalf("Failed to clear element with error %s", err)
	}
}
