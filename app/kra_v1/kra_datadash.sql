CREATE TABLE kra_tax_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    total_bets INT NOT NULL,
    total_stake DECIMAL(10, 2) NOT NULL,
    excise_duty_stake DECIMAL(10, 2) NOT NULL,
    excise_duty_paid DECIMAL(10, 2) NOT NULL,
    excise_duty_unpaid DECIMAL(10, 2) NOT NULL,
    total_winnings DECIMAL(10, 2) NOT NULL,
    totalwinning_bets DECIMAL(10, 2) NOT NULL,
    WHTOn_winnings DECIMAL(10, 2) NOT NULL,
    WHT_paid DECIMAL(10, 2) NOT NULL,
    WHT_unpaid DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
