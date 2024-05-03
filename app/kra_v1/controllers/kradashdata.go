package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

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

	validateKraData(data)

    // // Validate fetched data
	// if err := validateKraData(data); err != nil {
	// 	log.Printf("Failed to validate KRA data: %v", err)
	// 	return c.JSON(http.StatusBadRequest, map[string]interface{}{
	// 		"error": fmt.Sprintf("Validation error: %v", err),
	// 	})
	// }
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
func validateKraData(data *KraTaxData) {
	// Check if data is nil
	if data == nil {
		fmt.Println("KRA data is nil")
		return
	}

	// Validate specific fields in the KraTaxData struct
	if data.TotalBets <= 0 {
		fmt.Println("Invalid total bets: must be greater than zero, setting to zero")
		data.TotalBets = 0
	}

	if data.TotalStake <= 0 {
		fmt.Println("Invalid total stake: must be greater than zero, setting to zero")
		data.TotalStake = 0
	}

	if data.ExciseDutyUnpaid < 0 {
		fmt.Println("Invalid excise duty unpaid: must be non-negative, setting to zero")
		data.ExciseDutyUnpaid = 0
	}

	if data.ExciseDutyPaid < 0 {
		fmt.Println("Invalid excise duty paid: must be non-negative, setting to zero")
		data.ExciseDutyPaid = 0
	}

	if data.ExciseDutyStake < 0 {
		fmt.Println("Invalid excise duty stake: must be non-negative, setting to zero")
		data.ExciseDutyStake = 0
	}

	if data.TotalWinnings < 0 {
		fmt.Println("Invalid total winnings: must be non-negative, setting to zero")
		data.TotalWinnings = 0
	}

	if data.TotalWinningBets < 0 {
		fmt.Println("Invalid total winning bets: must be non-negative, setting to zero")
		data.TotalWinningBets = 0
	}

	if data.WHTOnWinning < 0 {
		fmt.Println("Invalid WHT on winning: must be non-negative, setting to zero")
		data.WHTOnWinning = 0
	}

	if data.WHTPaid < 0 {
		fmt.Println("Invalid WHT paid: must be non-negative, setting to zero")
		data.WHTPaid = 0
	}

	if data.WHTUnpaid < 0 {
		fmt.Println("Invalid WHT unpaid: must be non-negative, setting to zero")
		data.WHTUnpaid = 0
	}
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




