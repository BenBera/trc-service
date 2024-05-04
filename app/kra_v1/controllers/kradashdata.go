package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)


func (a *Api) GetKqraDashData(c echo.Context, db *sql.DB, redisCon *redis.Client) error {
    // Parse request payload
    u := new(KraTaxInfoWrapper)
    if err := c.Bind(u); err != nil {
        log.Printf("Failed to parse request body: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error": "Invalid parameters passed",
        })
    }

    // Fetch data based on the request payload
    data, err := a.fetchKraData()
    if err != nil {
        log.Printf("Failed to fetch KRA data: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "error": "Failed to fetch KRA data",
        })
    }


    // Validate fetched data
	if err := validateKraData(data); err != nil {
		log.Printf("Failed to validate KRA data: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

    // Store validated data in the database
    if err := insertKraData(db, data); err != nil {
        log.Printf("Failed to insert KRA data into database: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "error": "Failed to insert KRA data into database",
        })
    }

    // Return success response
    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "Request processed successfully",
        "data":    data,
    })
}


// validateKraData validates the fetched KRA data
func validateKraData(data *KraTaxData) error {
	if data == nil {
		return fmt.Errorf("validation error: KRA data is nil")
	}

	var errs []string

	if data.TotalBets <= 0 {
		data.TotalBets = 0
		errs = append(errs, "Invalid total bets: must be greater than zero")
	}

	if data.TotalStake <= 0 {
		data.TotalStake = 0
		errs = append(errs, "Invalid total stake: must be greater than zero")
	}

	if data.ExciseDutyUnpaid < 0 {
		data.ExciseDutyUnpaid = 0
		errs = append(errs, "Invalid excise duty unpaid: must be non-negative")
	}

	if data.ExciseDutyPaid < 0 {
		data.ExciseDutyPaid = 0
		errs = append(errs, "Invalid excise duty paid: must be non-negative")
	}

	if data.ExciseDutyStake < 0 {
		data.ExciseDutyStake = 0
		errs = append(errs, "Invalid excise duty stake: must be non-negative")
	}

	if data.TotalWinnings < 0 {
		data.TotalWinnings = 0
		errs = append(errs, "Invalid total winnings: must be non-negative")
	}

	if data.TotalWinningBets < 0 {
		data.TotalWinningBets = 0
		errs = append(errs, "Invalid total winning bets: must be non-negative")
	}

	if data.WHTOnWinning < 0 {
		data.WHTOnWinning = 0
		errs = append(errs, "Invalid WHT on winning: must be non-negative")
	}

	if data.WHTPaid < 0 {
		data.WHTPaid = 0
		errs = append(errs, "Invalid WHT paid: must be non-negative")
	}

	if data.WHTUnpaid < 0 {
		data.WHTUnpaid = 0
		errs = append(errs, "Invalid WHT unpaid: must be non-negative")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation error(s):\n%s", strings.Join(errs, "\n"))
	}

	return nil
}
 func insertKraData(db *sql.DB, data *KraTaxData) error {
    // Prepare the SQL INSERT statement
    query := `
        INSERT INTO kra_datadashdata (
            total_bets,
            total_stake,
            excise_duty_unpaid
			excise_duty_paid
			excise_duty_stake
			total_winnings
			total_winning_bets
			WHTOn_winning
			WHT_paid
			WHT_unpaid
            
        ) VALUES (?, ?, ?,?, ?, ?, ?, ?, ?, ?)
    `

    // Execute the SQL INSERT statement with the data parameters
    _, err := db.Exec(query,
        data.TotalBets,
        data.TotalStake,
        data.ExciseDutyUnpaid,
		data.ExciseDutyPaid,
		data.ExciseDutyStake,
		data.TotalWinnings,
		data.TotalWinningBets,
		data.WHTOnWinning,
		data.WHTPaid,
		data.WHTUnpaid,
    )
    if err != nil {
        return fmt.Errorf("failed to insert KRA data into database: %v", err)
    }

    return nil
}




