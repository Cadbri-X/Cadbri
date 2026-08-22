package engines

import (
	"cadbri/internal/engine"
)

// RegisterAll registers all built-in search engines and their shortcut bangs into the GlobalRegistry.
func RegisterAll() {
	r := engine.GlobalRegistry

	// 🌐 Category 1: Web Search Engines
	r.Register(NewGoogleEngine())
	r.RegisterShortcut("!g", "google")
	r.RegisterShortcut("!google", "google")

	r.Register(NewBingEngine())
	r.RegisterShortcut("!b", "bing")
	r.RegisterShortcut("!bing", "bing")

	r.Register(NewDuckDuckGoEngine())
	r.RegisterShortcut("!ddg", "duckduckgo")
	r.RegisterShortcut("!duckduckgo", "duckduckgo")

	r.Register(NewBraveEngine())
	r.RegisterShortcut("!brave", "brave")
	r.RegisterShortcut("!br", "brave")

	r.Register(NewWikipediaEngine())
	r.RegisterShortcut("!w", "wikipedia")
	r.RegisterShortcut("!wiki", "wikipedia")

	r.Register(NewYahooEngine())
	r.RegisterShortcut("!y", "yahoo")
	r.RegisterShortcut("!yahoo", "yahoo")

	r.Register(NewBaiduEngine())
	r.RegisterShortcut("!bd", "baidu")
	r.RegisterShortcut("!baidu", "baidu")

	r.Register(NewNaverEngine())
	r.RegisterShortcut("!nv", "naver")
	r.RegisterShortcut("!naver", "naver")

	r.Register(NewStartpageEngine())
	r.RegisterShortcut("!sp", "startpage")
	r.RegisterShortcut("!startpage", "startpage")

	r.Register(NewWibyEngine())
	r.RegisterShortcut("!wiby", "wiby")

	r.Register(NewQwantEngine())
	r.RegisterShortcut("!qw", "qwant")
	r.RegisterShortcut("!qwant", "qwant")

	r.Register(NewMojeekEngine())
	r.RegisterShortcut("!mj", "mojeek")
	r.RegisterShortcut("!mojeek", "mojeek")

	r.Register(NewSogouEngine())
	r.RegisterShortcut("!sogou", "sogou")

	r.Register(NewSo360Engine())
	r.RegisterShortcut("!360", "so360")
	r.RegisterShortcut("!so360", "so360")

	r.Register(NewSeznamEngine())
	r.RegisterShortcut("!sz", "seznam")
	r.RegisterShortcut("!seznam", "seznam")

	r.Register(NewPresearchEngine())
	r.RegisterShortcut("!pre", "presearch")
	r.RegisterShortcut("!presearch", "presearch")

	r.Register(NewMarginaliaEngine())
	r.RegisterShortcut("!marginalia", "marginalia")
	r.RegisterShortcut("!mar", "marginalia")

	r.Register(NewWebCrawlerEngine())
	r.RegisterShortcut("!webcrawler", "webcrawler")
	r.RegisterShortcut("!wc", "webcrawler")

	r.Register(NewExciteEngine())
	r.RegisterShortcut("!excite", "excite")
	r.RegisterShortcut("!ex", "excite")

	r.Register(NewMetaCrawlerEngine())
	r.RegisterShortcut("!metacrawler", "metacrawler")
	r.RegisterShortcut("!mc", "metacrawler")

	r.Register(NewEcosiaEngine())
	r.RegisterShortcut("!ecosia", "ecosia")
	r.RegisterShortcut("!eco", "ecosia")

	r.Register(NewSwisscowsEngine())
	r.RegisterShortcut("!swisscows", "swisscows")
	r.RegisterShortcut("!swiss", "swisscows")

	r.Register(NewYandexEngine())
	r.RegisterShortcut("!yandex", "yandex")
	r.RegisterShortcut("!ya", "yandex")

	r.Register(NewYepEngine())
	r.RegisterShortcut("!yep", "yep")

	r.Register(NewBlackleEngine())
	r.RegisterShortcut("!blackle", "blackle")
	r.RegisterShortcut("!blk", "blackle")

	r.Register(NewDogpileEngine())
	r.RegisterShortcut("!dogpile", "dogpile")
	r.RegisterShortcut("!dp", "dogpile")

	r.Register(NewAOLEngine())
	r.RegisterShortcut("!aol", "aol")

	r.Register(NewMailRuEngine())
	r.RegisterShortcut("!mailru", "mailru")
	r.RegisterShortcut("!mail", "mailru")

	// 💻 Category 2: Developer & IT
	r.Register(NewGitHubEngine())
	r.RegisterShortcut("!gh", "github")
	r.RegisterShortcut("!github", "github")

	r.Register(NewGitLabEngine())
	r.RegisterShortcut("!gl", "gitlab")
	r.RegisterShortcut("!gitlab", "gitlab")

	r.Register(NewGiteaEngine())
	r.RegisterShortcut("!cb", "codeberg")
	r.RegisterShortcut("!gitea", "codeberg")

	r.Register(NewNPMEngine())
	r.RegisterShortcut("!npm", "npm")

	r.Register(NewPyPIEngine())
	r.RegisterShortcut("!pypi", "pypi")
	r.RegisterShortcut("!pip", "pypi")

	r.Register(NewCratesEngine())
	r.RegisterShortcut("!crates", "crates")
	r.RegisterShortcut("!cargo", "crates")

	r.Register(NewRubyGemsEngine())
	r.RegisterShortcut("!gem", "rubygems")
	r.RegisterShortcut("!rubygems", "rubygems")

	r.Register(NewPkgGoDevEngine())
	r.RegisterShortcut("!go", "pkg_go_dev")
	r.RegisterShortcut("!godoc", "pkg_go_dev")

	r.Register(NewStackExchangeEngine())
	r.RegisterShortcut("!so", "stackoverflow")
	r.RegisterShortcut("!stack", "stackoverflow")

	// 🎓 Category 3: Academic & Science
	r.Register(NewGoogleScholarEngine())
	r.RegisterShortcut("!scholar", "google_scholar")
	r.RegisterShortcut("!gs", "google_scholar")

	r.Register(NewArXivEngine())
	r.RegisterShortcut("!arxiv", "arxiv")
	r.RegisterShortcut("!ax", "arxiv")

	r.Register(NewPubMedEngine())
	r.RegisterShortcut("!pubmed", "pubmed")
	r.RegisterShortcut("!pm", "pubmed")

	r.Register(NewCrossrefEngine())
	r.RegisterShortcut("!crossref", "crossref")
	r.RegisterShortcut("!doi", "crossref")

	r.Register(NewSemanticScholarEngine())
	r.RegisterShortcut("!semantic", "semantic_scholar")
	r.RegisterShortcut("!ss", "semantic_scholar")

	r.Register(NewOpenAlexEngine())
	r.RegisterShortcut("!openalex", "openalex")
	r.RegisterShortcut("!oa", "openalex")

	r.Register(NewADSEngine())
	r.RegisterShortcut("!ads", "ads")

	// 🎬 Category 4: Videos & Music
	r.Register(NewYouTubeEngine())
	r.RegisterShortcut("!yt", "youtube")
	r.RegisterShortcut("!youtube", "youtube")

	r.Register(NewVimeoEngine())
	r.RegisterShortcut("!vimeo", "vimeo")
	r.RegisterShortcut("!vm", "vimeo")

	r.Register(NewDailymotionEngine())
	r.RegisterShortcut("!dm", "dailymotion")
	r.RegisterShortcut("!dailymotion", "dailymotion")

	r.Register(NewBilibiliEngine())
	r.RegisterShortcut("!bili", "bilibili")
	r.RegisterShortcut("!bilibili", "bilibili")

	r.Register(NewSoundCloudEngine())
	r.RegisterShortcut("!sc", "soundcloud")
	r.RegisterShortcut("!soundcloud", "soundcloud")

	r.Register(NewDeezerEngine())
	r.RegisterShortcut("!dz", "deezer")
	r.RegisterShortcut("!deezer", "deezer")

	r.Register(NewRadioBrowserEngine())
	r.RegisterShortcut("!radio", "radio_browser")

	// 🖼️ Category 5: Images & Art
	r.Register(NewUnsplashEngine())
	r.RegisterShortcut("!unsplash", "unsplash")
	r.RegisterShortcut("!un", "unsplash")

	r.Register(NewPexelsEngine())
	r.RegisterShortcut("!pexels", "pexels")
	r.RegisterShortcut("!px", "pexels")

	r.Register(NewPinterestEngine())
	r.RegisterShortcut("!pin", "pinterest")
	r.RegisterShortcut("!pinterest", "pinterest")

	r.Register(NewDeviantArtEngine())
	r.RegisterShortcut("!da", "deviantart")
	r.RegisterShortcut("!deviantart", "deviantart")

	r.Register(NewWallhavenEngine())
	r.RegisterShortcut("!wh", "wallhaven")
	r.RegisterShortcut("!wallhaven", "wallhaven")

	r.Register(NewWww1xEngine())
	r.RegisterShortcut("!1x", "1x")

	// 🗺️ Category 6: Maps & Locations
	r.Register(NewOpenStreetMapEngine())
	r.RegisterShortcut("!osm", "openstreetmap")
	r.RegisterShortcut("!map", "openstreetmap")

	r.Register(NewAppleMapsEngine())
	r.RegisterShortcut("!am", "apple_maps")
	r.RegisterShortcut("!applemaps", "apple_maps")

	r.Register(NewPhotonEngine())
	r.RegisterShortcut("!photon", "photon")
	r.RegisterShortcut("!geo", "photon")

	// 📰 Category 7: News
	r.Register(NewGoogleNewsEngine())
	r.RegisterShortcut("!gn", "google_news")
	r.RegisterShortcut("!gnews", "google_news")

	r.Register(NewBingNewsEngine())
	r.RegisterShortcut("!bn", "bing_news")
	r.RegisterShortcut("!bnews", "bing_news")

	r.Register(NewYahooNewsEngine())
	r.RegisterShortcut("!yn", "yahoo_news")
	r.RegisterShortcut("!ynews", "yahoo_news")

	r.Register(NewReutersEngine())
	r.RegisterShortcut("!reuters", "reuters")
	r.RegisterShortcut("!rt", "reuters")

	r.Register(NewTagesschauEngine())
	r.RegisterShortcut("!ts", "tagesschau")
	r.RegisterShortcut("!tagesschau", "tagesschau")
}
