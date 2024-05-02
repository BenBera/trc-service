package controllers


import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// GetKraDashData handles the KRA dashboard data processing
func (a *Api) GetKraDashData(c echo.Context) error {
    u := new(KraTaxInfoWrapper)

    // Bind request payload to u
    if err := c.Bind(u); err != nil {
        log.Printf("Failed to parse request body: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error": "Invalid parameters passed",
        })
    }

    // Process KRA dashboard data
    err := a.ProcessKraDashboardData(a.DB, u, c)
    if err != nil {
        log.Printf("Failed to process KRA dashboard data: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "error": "Failed to process KRA dashboard data",
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "Request received and processed successfully",
        "payload": u,
    })
}


// ProcessKraDashboardData processes the KRA dashboard data and inserts into the database
func (a *Api)ProcessKraDashboardData(db *sql.DB, u *KraTaxInfoWrapper, c echo.Context) error {
	// Prepare parameters for signature calculation
	KraParams := []string{
		fmt.Sprintf("total_bets=%d", u.TotalBets),
		fmt.Sprintf("total_stake=%.2f", u.TotalStake),
		fmt.Sprintf("excise_duty_stake=%.2f", u.ExciseDutyStake),
		fmt.Sprintf("excise_duty_paid=%.2f", u.ExciseDutyPaid),
		fmt.Sprintf("excise_duty_unpaid=%.2f", u.ExciseDutyUnpaid),
		fmt.Sprintf("total_winnings=%.2f", u.TotalWinnings),
		fmt.Sprintf("totalwinning_bets=%.2f", u.TotalWinningBets),
		fmt.Sprintf("WHTOn_winnings=%.2f", u.WHTOnWinning),
		fmt.Sprintf("WHT_paid=%.2f", u.WHTPaid),
		fmt.Sprintf("WHT_unpaid=%.2f", u.WHTUnpaid),
	}

	// Calculate signature 	and Extract signature from payload

	privateKey := GetKeyWithDefault(a.Config, "kradata", "private-key", "")
	signature := calculateSHA256(strings.Join(KraParams, "") + privateKey)


	// Verify signature
	if signature != u.PassKey {
		return RespondRaw(c, http.StatusInternalServerError, "Signature mismatched")
	}



	// Log request payload
	logRequestPayload(u)

	// Insert data into the database
	err := InsertKraTaxData(db, u)
	if err != nil {
		return fmt.Errorf("failed to insert data into database: %v", err)
	}

	return nil
}

func InsertKraTaxData(db *sql.DB, u *KraTaxInfoWrapper) error {
	query := `
		INSERT INTO kra_tax_data (
			total_bets,
			total_stake,
			excise_duty_stake,
			excise_duty_paid,
			excise_duty_unpaid,
			total_winnings,
			totalwinning_bets,
			WHTOn_winnings,
			WHT_paid,
			WHT_unpaid
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query,
		u.TotalBets,
		u.TotalStake,
		u.ExciseDutyStake,
		u.ExciseDutyPaid,
		u.ExciseDutyUnpaid,
		u.TotalWinnings,
		u.TotalWinningBets,
		u.WHTOnWinning,
		u.WHTPaid,
		u.WHTUnpaid,
	)
	if err != nil {
		return fmt.Errorf("failed to insert data into database: %v", err)
	}

	return nil
}

func logRequestPayload(u *KraTaxInfoWrapper) {
	js, err := json.Marshal(u)
	if err != nil {
		log.Printf("Failed to marshal payload to JSON: %v", err)
	} else {
		log.Printf("Request payload: %s", string(js))
	}
}

func (a *Api) handleGetKraDashData(c echo.Context) echo.HandlerFunc {
    return func(c echo.Context) error {
        return a.GetKraDashData(c) 
    }
}


