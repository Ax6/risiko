package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	risk "github.com/Ax6/risk/pkg/risk"
)

func saveCSV(filename string, header []string, data [][]string) error {
	// Create or open the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)

	// Write the header
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Flush the writer
	writer.Flush()

	// Check for errors while writing
	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}

func generateCSVTables(simResult risk.SimulationSweep, unitsSweep int) {
	// Initialize slices to store the data for the two tables
	var victoryTable [][]string
	var attackersLeftTable [][]string

	// Create headers for the tables
	header := []string{"nUnits"}
	for i := 1; i <= unitsSweep; i++ {
		header = append(header, strconv.Itoa(i)) // Add attacker units to header
	}

	// Iterate over the simulation results to calculate percentages
	for nDefenders := 1; nDefenders <= unitsSweep; nDefenders++ {
		victoryRow := []string{strconv.Itoa(nDefenders)}       // First column: nDefenderUnits
		attackersLeftRow := []string{strconv.Itoa(nDefenders)} // First column: nDefenderUnits

		// Iterate over attacker units (columns)
		for nAttackers := 1; nAttackers <= unitsSweep; nAttackers++ {
			result := simResult[nAttackers][nDefenders]

			// Calculate the victory percentage for the attacker
			victoryPercentage := float64(result.NAttackerWon) / float64(result.NRuns)
			// Add the result as a string to the victory table row
			victoryRow = append(victoryRow, fmt.Sprintf("%.6f", victoryPercentage))

			// Calculate the percentage of attackers left
			attackersLeftPercentage := float64(result.TotalAttackerUnitsLeft) / float64(result.NAttackerWon)
			// Add the result as a string to the attackers left table row
			attackersLeftRow = append(attackersLeftRow, fmt.Sprintf("%.6f", attackersLeftPercentage))
		}

		// Add the rows to the tables
		victoryTable = append(victoryTable, victoryRow)
		attackersLeftTable = append(attackersLeftTable, attackersLeftRow)
	}

	// Save the victory table to a CSV file
	if err := saveCSV("victory_percentage.csv", header, victoryTable); err != nil {
		log.Fatalf("Error saving victory table: %v", err)
	}

	// Save the attackers left table to a CSV file
	if err := saveCSV("attackers_left_percentage.csv", header, attackersLeftTable); err != nil {
		log.Fatalf("Error saving attackers left table: %v", err)
	}

	log.Println("CSV files created successfully.")
}

func main() {
	// Context for simulation
	ctx := context.Background()

	// Sample parameters
	nRuns := 10000
	unitsSweep := 20
	attacker := risk.NewMaxAttackersStrategy(risk.FairDicesGen)
	defender := risk.NewMaxDefendersStrategy(risk.FairDicesGen)
	log.Println("Starting simulation....")

	// Simulate and get the results
	simResult, err := risk.Simulate(ctx, nRuns, unitsSweep, attacker, defender)
	if err != nil {
		log.Fatalf("Error in simulation: %v", err)
	}
	log.Println("Simulation finished successfully!")

	// Generate and save the CSV tables
	generateCSVTables(simResult, unitsSweep)
}