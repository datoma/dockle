package manifest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/goodwithtech/dockle/pkg/types"
)

func TestAssess(t *testing.T) {
	var tests = map[string]struct {
		path     string
		assesses []*types.Assessment
	}{
		"RootDefault": {
			path: "./testdata/root_default.json",
			assesses: []*types.Assessment{
				{
					Type:     types.AvoidRootDefault,
					Filename: "docker config",
				},
				{
					Type:     types.AddHealthcheck,
					Filename: "docker config",
				},
			},
		},
		"ApkCached": {
			path: "./testdata/apk_cache.json",

			assesses: []*types.Assessment{
				{
					Type:     types.AvoidRootDefault,
					Filename: "docker config",
				},
				{
					Type:     types.UseApkAddNoCache,
					Filename: "docker config",
				},
				{
					Type:     types.AddHealthcheck,
					Filename: "docker config",
				},
				{
					Type:     types.UseCOPY,
					Filename: "docker config",
				},
			},
		},
	}

	for testname, v := range tests {
		d, err := loadImageFromFile(v.path)

		if err != nil {
			t.Errorf("%s : can't open file %s", testname, v.path)
			continue
		}
		actual, err := checkAssessments(d)
		if err != nil {
			t.Errorf("%s : catch the error : %v", testname, err)
		}

		diff, equal := messagediff.PrettyDiff(
			sortByType(v.assesses),
			sortByType(actual),
			messagediff.IgnoreStructField("Desc"),
		)
		if !equal {
			t.Errorf("%s diff : %v", testname, diff)
		}
	}
}

func TestSplitByCommands(t *testing.T) {
	var tests = map[string]struct {
		path     string
		index    int
		expected map[int][]string
	}{
		"RootDefault": {
			path:  "./testdata/root_default.json",
			index: 1,
			expected: map[int][]string{
				0: {"/bin/sh", "-c", "#(nop)", "CMD", "[\"/bin/sh\"]"},
			},
		},
		"Nginx": {
			path:  "./testdata/nginx.json",
			index: 6,
			expected: map[int][]string{
				0:  {"/bin/sh", "-c", "set", "-x"},
				1:  {"addgroup", "--system", "--gid", "101", "nginx"},
				2:  {"adduser", "--system", "--disabled-login", "--ingroup", "nginx", "--no-create-home", "--home", "/nonexistent", "--gecos", "\"nginx", "user\"", "--shell", "/bin/false", "--uid", "101", "nginx"},
				3:  {"apt-get", "update"},
				4:  {"apt-get", "install", "--no-install-recommends", "--no-install-suggests", "-y", "gnupg1", "apt-transport-https", "ca-certificates"},
				5:  {"\tNGINX_GPGKEY=573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62;", "\tfound='';", "\tfor", "server", "in", "\t\tha.pool.sks-keyservers.net", "\t\thkp://keyserver.ubuntu.com:80", "\t\thkp://p80.pool.sks-keyservers.net:80", "\t\tpgp.mit.edu", "\t;", "do", "\t\techo", "\"Fetching", "GPG", "key", "$NGINX_GPGKEY", "from", "$server\";", "\t\tapt-key", "adv", "--keyserver", "\"$server\"", "--keyserver-options", "timeout=10", "--recv-keys", "\"$NGINX_GPGKEY\""},
				6:  {"found=yes"},
				7:  {"break;", "\tdone;", "\ttest", "-z", "\"$found\""},
				8:  {"echo", ">&2", "\"error:", "failed", "to", "fetch", "GPG", "key", "$NGINX_GPGKEY\""},
				9:  {"exit", "1;", "\tapt-get", "remove", "--purge", "--auto-remove", "-y", "gnupg1"},
				10: {"rm", "-rf", "/var/lib/apt/lists/*"},
				11: {"dpkgArch=\"$(dpkg", "--print-architecture)\""},
				12: {"nginxPackages=\"", "\t\tnginx=${NGINX_VERSION}-${PKG_RELEASE}", "\t\tnginx-module-xslt=${NGINX_VERSION}-${PKG_RELEASE}", "\t\tnginx-module-geoip=${NGINX_VERSION}-${PKG_RELEASE}", "\t\tnginx-module-image-filter=${NGINX_VERSION}-${PKG_RELEASE}", "\t\tnginx-module-njs=${NGINX_VERSION}.${NJS_VERSION}-${PKG_RELEASE}", "\t\""},
				13: {"case", "\"$dpkgArch\"", "in", "\t\tamd64|i386)", "\t\t\techo", "\"deb", "https://nginx.org/packages/mainline/debian/", "stretch", "nginx\"", ">>", "/etc/apt/sources.list.d/nginx.list"},
				14: {"apt-get", "update", "\t\t\t;;", "\t\t*)", "\t\t\techo", "\"deb-src", "https://nginx.org/packages/mainline/debian/", "stretch", "nginx\"", ">>", "/etc/apt/sources.list.d/nginx.list"},
				15: {"tempDir=\"$(mktemp", "-d)\""},
				16: {"chmod", "777", "\"$tempDir\""},
				17: {"savedAptMark=\"$(apt-mark", "showmanual)\""},
				18: {"apt-get", "update"},
				19: {"apt-get", "build-dep", "-y", "$nginxPackages"},
				20: {"(", "\t\t\t\tcd", "\"$tempDir\""},
				21: {"DEB_BUILD_OPTIONS=\"nocheck", "parallel=$(nproc)\"", "\t\t\t\t\tapt-get", "source", "--compile", "$nginxPackages", "\t\t\t)"},
				22: {"apt-mark", "showmanual", "|", "xargs", "apt-mark", "auto", ">", "/dev/null"},
				23: {"{", "[", "-z", "\"$savedAptMark\"", "]", "||", "apt-mark", "manual", "$savedAptMark;", "}"},
				24: {"ls", "-lAFh", "\"$tempDir\""},
				25: {"(", "cd", "\"$tempDir\""},
				26: {"dpkg-scanpackages", ".", ">", "Packages", ")"},
				27: {"grep", "'^Package:", "'", "\"$tempDir/Packages\""},
				28: {"echo", "\"deb", "[", "trusted=yes", "]", "file://$tempDir", "./\"", ">", "/etc/apt/sources.list.d/temp.list"},
				29: {"apt-get", "-o", "Acquire::GzipIndexes=false", "update", "\t\t\t;;", "\tesac"},
				30: {"apt-get", "install", "--no-install-recommends", "--no-install-suggests", "-y", "\t\t\t\t\t\t$nginxPackages", "\t\t\t\t\t\tgettext-base"},
				31: {"apt-get", "remove", "--purge", "--auto-remove", "-y", "apt-transport-https", "ca-certificates"},
				32: {"rm", "-rf", "/var/lib/apt/lists/*", "/etc/apt/sources.list.d/nginx.list"},
				33: {"if", "[", "-n", "\"$tempDir\"", "];", "then", "\t\tapt-get", "purge", "-y", "--auto-remove"},
				34: {"rm", "-rf", "\"$tempDir\"", "/etc/apt/sources.list.d/temp.list;", "\tfi"},
			},
		},
	}

	for testname, v := range tests {
		d, err := loadImageFromFile(v.path)
		if err != nil {
			t.Errorf("%s : can't open file %s", testname, v.path)
			continue
		}
		cmd := d.History[v.index]
		actual := splitByCommands(cmd.CreatedBy)
		diff, equal := messagediff.PrettyDiff(
			v.expected,
			actual,
		)
		if !equal {
			t.Errorf("%s diff : %v", testname, diff)
		}
	}
}

func TestReducableApkAdd(t *testing.T) {
	var tests = map[string]struct {
		cmdSlices map[int][]string
		expected  bool
	}{
		"Reducable": {
			cmdSlices: map[int][]string{
				0: {
					"apk", "add", "git",
				},
			},
			expected: true,
		},
		"UnReducable": {
			cmdSlices: map[int][]string{
				0: {
					"apk", "add", "--no-cache", "git",
				},
			},
			expected: false,
		},
	}
	for testname, v := range tests {
		actual := reducableApkAdd(v.cmdSlices)
		if actual != v.expected {
			t.Errorf("%s want: %t, got %t", testname, v.expected, actual)
		}
	}
}

func TestReducableAptGetUpdate(t *testing.T) {
	var tests = map[string]struct {
		cmdSlices map[int][]string
		expected  bool
	}{
		"Reducable": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "update",
				},
				1: {
					"apt-get", "purge",
				},
			},
			expected: true,
		},
		"NoUpdate": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "install",
				},
				1: {
					"apt-get", "purge",
				},
			},
			expected: false,
		},
		"UnReducable": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "update",
				},
				1: {
					"apt-get", "-y", "--no-install-recommends", "install",
				},
			},
			expected: false,
		},
	}
	for testname, v := range tests {
		actual := reducableAptGetUpdate(v.cmdSlices)
		if actual != v.expected {
			t.Errorf("%s want: %t, got %t", testname, v.expected, actual)
		}
	}
}

func TestReducableAptGetInstall(t *testing.T) {
	var tests = map[string]struct {
		cmdSlices map[int][]string
		expected  bool
	}{
		"Reducable": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "-y", "install",
				},
				1: {
					"apt-get", "update",
				},
			},
			expected: true,
		},
		"NoInstall": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "update",
				},
				1: {
					"apt-get", "purge",
				},
			},
			expected: false,
		},
		"UnReducable": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "install",
				},
				1: {
					"rm", "-fR", "/var/lib/apt/lists/*",
				},
			},
			expected: false,
		},
		"UnReducable2": {
			cmdSlices: map[int][]string{
				0: {
					"apt-get", "install",
				},
				1: {
					"rm", "-rf", "/var/lib/apt/lists",
				},
			},
			expected: false,
		},
	}
	for testname, v := range tests {
		actual := reducableAptGetInstall(v.cmdSlices)
		if actual != v.expected {
			t.Errorf("%s want: %t, got %t", testname, v.expected, actual)
		}
	}
}

func loadImageFromFile(path string) (config types.Image, err error) {
	read, err := os.Open(path)
	if err != nil {
		return config, err
	}
	filebytes, err := ioutil.ReadAll(read)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(filebytes, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func sortByType(assesses []*types.Assessment) []*types.Assessment {
	sort.Slice(assesses, func(i, j int) bool {
		if assesses[i].Type != assesses[j].Type {
			return assesses[i].Type < assesses[j].Type
		}
		return assesses[i].Type < assesses[j].Type
	})
	return assesses
}
