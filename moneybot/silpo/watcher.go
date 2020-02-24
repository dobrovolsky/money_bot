package silpo

import (
	"encoding/json"
	mb "github.com/dobrovolsky/money_bot/moneybot"
	"github.com/sirupsen/logrus"
	"time"
)

type ResponseDataList struct {
	SumBalance []struct {
		Headers []Identities `json:"headers"`
	} `json:"sumBalance"`
}

type ChequeLines struct {
	LagerNameUA string  `json:"lagerNameUA"`
	SumLine     float64 `json:"sumLine"`
}
type ResponseDataDetail struct {
	ChequesInfos []struct {
		ChequeLines []ChequeLines `json:"chequeLines"`
	} `json:"chequesInfos"`
}

func Watcher(accessToken string, refreshToken string, events chan mb.Item) {
	c := Client{}

	c.Tokens.AccessToken.Value = accessToken
	c.Tokens.RefreshToken.Value = refreshToken

	lastLogged := getLastLogged()

	for {
		// TODO: handle token expiration
		data, err := c.GetLastChequeHeaders()
		if err != nil {
			logrus.Error(err)
			return
		}

		response := ResponseDataList{}
		err = json.Unmarshal(data, &response)
		if err != nil {
			logrus.Error(err)
			return
		}

		newItems := getNewCheques(lastLogged, response)
		logrus.Infof("new items since %s: %v", lastLogged, newItems)

		for _, info := range newItems {
			data, err := c.GetChequesInfos(info.ChequeID, info.Created, info.FilID)
			if err != nil {
				logrus.Error(err)
				return
			}

			response := ResponseDataDetail{}
			err = json.Unmarshal(data, &response)
			if err != nil {
				logrus.Error(err)
				return
			}

			logrus.Infof("new items for detailed %s: %v", info.Created, response.ChequesInfos[0])

			for _, line := range response.ChequesInfos[0].ChequeLines {
				item := mb.Item{Name: line.LagerNameUA, Amount: line.SumLine}
				events <- item
				// TODO: update lastLogged in db

				lastLogged = info.Created
			}
		}

		// TODO: add time to config
		time.Sleep(60 * time.Second)
	}
}

func getNewCheques(lastLogged string, data ResponseDataList) []Identities {
	var newUpdates []Identities

	for _, sumBalance := range data.SumBalance {
		for _, header := range sumBalance.Headers {
			if lastLogged < header.Created {
				newUpdates = append(newUpdates, header)
			}
		}
	}

	return newUpdates
}

func getLastLogged() string {
	// TODO: add db integration
	return "2020-02-23T18:13:33"
}
