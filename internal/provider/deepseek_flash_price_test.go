package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
)

// deepseek-v4.1-flash is priced at deepseek-flash, its maker's own
// listing, whatever canonical_model_id models.dev gives that entry: the
// canonical was deepseek/deepseek-v4.1-flash when #1520 priced the id by
// it, and models.dev has since moved it to deepseek/deepseek-v4.1-flash-1017,
// which left a vendor serving the id as given — WorkBuddy's subscription —
// unpriced again (hlyx1). As models.dev's deepseek entry lists it
// (2026-11-01), snapshot suffix and all.
func TestDeepSeekFlashAliasPrice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755)
	os.WriteFile(catalog.CachePath(), []byte(`{
	  "deepseek":{"id":"deepseek","models":{
	    "deepseek-flash":{"id":"deepseek-flash","cost":{"input":0.15,"output":0.6,"cache_read":0.003},"canonical_model_id":"deepseek/deepseek-v4.1-flash-1017"},
	    "deepseek-v4-flash":{"id":"deepseek-v4-flash","status":"deprecated","cost":{"input":0.15,"output":0.6,"cache_read":0.003},"canonical_model_id":"deepseek/deepseek-v4.1-flash-1017"},
	    "deepseek-v4-pro":{"id":"deepseek-v4-pro","cost":{"input":0.66,"output":1.98,"cache_read":0.022},"canonical_model_id":"deepseek/deepseek-v4-pro-0813"}}},
	  "aihubmix":{"id":"aihubmix","models":{
	    "deepseek-v4.1-flash":{"id":"deepseek-v4.1-flash","cost":{"input":0.155,"output":0.62,"cache_read":0.0031}}}}}`), 0o644)
	catalog.Reset()
	t.Cleanup(catalog.Reset)

	for _, id := range []string{"deepseek-v4.1-flash", "deepseek/deepseek-v4.1-flash", "DeepSeek-V4.1-Flash"} {
		pr, ok := MakerPrice(id)
		if !ok || pr.Input != 0.15 || pr.Output != 0.6 || pr.CacheRead != 0.003 {
			t.Errorf("MakerPrice(%q) = %+v, %v; want deepseek-flash's $0.15/$0.6/$0.003", id, pr, ok)
		}
		// a subscription with no catalog of its own is priced at it too
		if pr, ok := (Provider{ID: "workbuddy", Name: "WorkBuddy"}).ListPrice(id); !ok || pr.Input != 0.15 {
			t.Errorf("workbuddy ListPrice(%q) = %+v, %v; want $0.15", id, pr, ok)
		}
	}
	// a vendor listing the id as given is asked before any alias
	aihubmix := Provider{ID: "aihubmix", Name: "AiHubMix", Catalog: "aihubmix"}
	if pr, ok := aihubmix.ListPrice("deepseek-v4.1-flash"); !ok || pr.Input != 0.155 {
		t.Errorf("aihubmix ListPrice(deepseek-v4.1-flash) = %+v, %v; want its own $0.155", pr, ok)
	}
	// no alias beyond this one id: the maker's own listed ids are priced as
	// themselves, a near-miss of the id stays unpriced
	for _, id := range []string{"deepseek-v4.1-pro", "deepseek-v4.1-flashx"} {
		if pr, ok := MakerPrice(id); ok {
			t.Errorf("MakerPrice(%q) = %+v; want unpriced", id, pr)
		}
	}
	if got := PricedName("deepseek-v4.1-flash"); got != "deepseek-flash" {
		t.Errorf("PricedName(deepseek-v4.1-flash) = %q", got)
	}
	if got := PricedName("deepseek-flash"); got != "deepseek-flash" {
		t.Errorf("PricedName(deepseek-flash) = %q", got)
	}
}
