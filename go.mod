module github.com/snipem/maniacforum

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/gizak/termui/v3 v3.1.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/stretchr/testify v1.5.1
)

// Use this fork for exposed TopRow attribute as long as this pull request is not merged: https://github.com/gizak/termui/pull/243
replace github.com/gizak/termui/v3 v3.1.1 => github.com/artyl/termui/v3 v3.1.1-0.20190829181539-c1256862160b
