package scheme

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"regexp"
	"strconv"

	"github.com/Laky-64/http"
)

func (ctx *Client) RefreshReleasedLayers() (latestVersion int, forceNoUpdate bool, err error) {
	res, err := http.ExecuteRequest(fmt.Sprintf("%s/schema", consts.MainReleasedTL))
	if err != nil {
		return 0, false, err
	}
	var versionsAvailable []int
	if rgx := regexp.MustCompile(`<li><a href="\?layer=([0-9]+)">`).FindAllStringSubmatch(res.String(), -1); len(rgx) > 0 {
		for _, v := range rgx {
			parsedLayer, _ := strconv.Atoi(v[1])
			versionsAvailable = append(versionsAvailable, parsedLayer)
		}
	}
	if len(versionsAvailable) == 0 {
		return 0, false, errors.New("failed to get the latest version of the TL scheme")
	}
	versionsAvailable = versionsAvailable[1:]
	startLayer := versionsAvailable[0]
	known, err := ctx.db.ReleasedStore.GetMaxReleasedLayer()
	if err != nil {
		return 0, false, err
	}
	if known == 0 {
		forceNoUpdate = true
	} else {
		startLayer = int(known)
	}
	latestVersion = versionsAvailable[len(versionsAvailable)-1]
	if startLayer < latestVersion {
		for _, layer := range versionsAvailable {
			if layer < startLayer {
				continue
			}
			tlRes, reqErr := http.ExecuteRequest(
				fmt.Sprintf("%s/schema/json", consts.MainReleasedTL),
				http.Cookies(map[string]string{
					"stel_dev_layer": strconv.Itoa(layer),
				}),
			)
			if reqErr != nil {
				return 0, false, reqErr
			}
			var releasedLayer types.ReleasedLayer
			if reqErr = json.Unmarshal(tlRes.Body, &releasedLayer); reqErr != nil {
				return 0, false, reqErr
			}
			if reqErr = ctx.db.ReplaceReleasedLayer(layer, releasedLayer); reqErr != nil {
				return 0, false, reqErr
			}
		}
	}
	return latestVersion, forceNoUpdate, nil
}
