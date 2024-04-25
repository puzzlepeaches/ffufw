package cmd

type Wordlist struct {
	Name string
	URL  string
}

var MiscWordlists = []Wordlist{
	// {
	// 	Name: "leaky-paths",
	// 	URL:  "https://raw.githubusercontent.com/ayoubfathi/leaky-paths/main/leaky-paths.txt",
	// },
	// {
	// 	Name: "assetnote-txt",
	// 	URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_txt_2023_10_28.txt",
	// },
	// {
	// 	Name: "assetnote-xml",
	// 	URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_xml_2023_10_28.txt",
	// },
	// {
	// 	Name: "assetnote-bak",
	// 	URL:  "https://wordlists-cdn.assetnote.io/data/manual/bak.txt",
	// },
	{
		Name: "seclists-logins",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/Logins.fuzz.txt",
	},
	{
		Name: "raft-medium-directories",
		URL:  "https://wordlists-cdn.assetnote.io/data/manual/raft-medium-directories-lowercase.txt",
	},
	{
		Name: "raft-medium-files",
		URL:  "https://wordlists-cdn.assetnote.io/data/manual/raft-medium-files-lowercase.txt",
	},
	{
		Name: "raft-large-words",
		URL:  "https://wordlists-cdn.assetnote.io/data/manual/raft-large-words.txt",
	},
}

var IisWordlists = []Wordlist{
	{
		Name: "assetnote-iis-auto",
		URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_aspx_asp_cfm_svc_ashx_asmx_2023_10_28.txt",
	},
	{
		Name: "chameleon-iis",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/IIS-ASP.txt",
	},
}

var PhpWordlists = []Wordlist{
	{
		Name: "assetnote-php-auto",
		URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_php_2023_10_28.txt",
	},
	{
		Name: "chameleon-php",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/PHP-Nuke.txt",
	},
}

var JavaWordlists = []Wordlist{
	{
		Name: "chameleon-java",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Java.txt",
	},
	{
		Name: "chameleon-spring",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Spring.txt",
	},
	{
		Name: "assetnote-jsp-auto",
		URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_jsp_jspa_do_action_2023_10_28.txt",
	},
	{
		Name: "assetnote-jsp-manual",
		URL:  "https://wordlists-cdn.assetnote.io/data/manual/jsp.txt",
	},
	{
		Name: "seclists-java-servlets",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/JavaServlets-Common.fuzz.txt",
	},
}

var ApiWordlists = []Wordlist{
	{
		Name: "assetnote-api-auto",
		URL:  "https://wordlists-cdn.assetnote.io/data/automated/httparchive_apiroutes_2023_10_28.txt",
	},
	{
		Name: "seclists-api",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/api/api-endpoints.txt",
	},
	{
		Name: "chameleon-spring-api",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Spring.txt",
	},
}

var PythonWordlists = []Wordlist{
	{
		Name: "chameleon-django",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Django.txt",
	},
	{
		Name: "chameleon-flask",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Flask.txt",
	},
}

var NginxWordlists = []Wordlist{
	{
		Name: "chameleon-nginx",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/nginx.txt",
	},
}

var SapWordlists = []Wordlist{
	{
		Name: "chameleon-sap",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/SAP.txt",
	},
	{
		Name: "seclists-sap-cloud",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/sap-analytics-cloud.txt",
	},
}

var RubyWordlists = []Wordlist{
	{
		Name: "chameleon-ruby",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/Ruby.txt",
	},
	{
		Name: "scumdestroy-rails",
		URL:  "https://gist.githubusercontent.com/scumdestroy/1f2bb7e2f5b80088cc934725bed7446d/raw/c2643ff64a6cd3a4af59375cb5b0239484cb62d2/ruby-on-rails-overdose.txt",
	},
}

var AdobeWordlists = []Wordlist{
	{
		Name: "chameleon-aem",
		URL:  "https://raw.githubusercontent.com/iustin24/chameleon/master/wordlists/AEM.txt",
	},
	{
		Name: "seclists-adobe-xml",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/AdobeXML.fuzz.txt",
	},
	{
		Name: "seclists-aem2",
		URL:  "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/aem2.txt",
	},
}
