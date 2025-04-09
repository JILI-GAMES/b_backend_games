package superace_deluxe

import (
	"fmt"
	"log"
	
	"github.com/gofiber/fiber/v2"
)

func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
    var rowReq RowBasedSpinRequest
    if err := c.BodyParser(&rowReq); err != nil {
        log.Printf("Error parsing row-based request: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  400,
            "message": "Invalid request format",
        })
    }
    
    // Validate request parameters
    if rowReq.BetAmount < 0.5 || (rowReq.Action != "SPIN" && rowReq.Action != "TRANSFORM") {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  400,
            "message": "Invalid bet amount or action",
        })
    }

    // Create a game state using the internal reel-based format
    gs := NewGameState(rowReq.BetAmount, rowReq.Game.Mode)
    gs.FreeSpins = rowReq.FreeSpins
    
    // Only set ComboMultiplier and convert Cards for TRANSFORM actions
    if rowReq.Action == "TRANSFORM" {
        gs.ComboMultiplier = rowReq.ComboMultiplier
        // Convert row-based cards from request to reel-based format for internal processing
        gs.Cards = ConvertToReelBased(rowReq.Cards)
    }

    // Use the shared settings service
    rtp, err := rg.Settings.GetRTP(rowReq.ClientID, rowReq.Game.ID, rowReq.PlayerID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  500,
            "message": "Failed to fetch RTP",
        })
    }

    fmt.Printf("RTP: %v\n", rtp)

    // Generate grid and calculate potential wins inside Spin/Transform
    var potentialWins float64
    if rowReq.Action == "SPIN" {
        potentialWins = gs.Spin()
    } else if rowReq.Action == "TRANSFORM" {
        potentialWins = gs.Transform()
    }

    payoutMultiplier := potentialWins / rowReq.BetAmount
    fmt.Printf("Potential wins: %v, Payout multiplier: %v\n", potentialWins, payoutMultiplier)

    // Use the shared RNG service
    rngResp, err := rg.RNG.GetOutcome(rowReq.ClientID, rowReq.Game.ID, rowReq.PlayerID, rtp, payoutMultiplier, rowReq.BetAmount)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  500,
            "message": "Failed to fetch RNG outcome",
        })
    }

    fmt.Printf("RNG response: {pref_outcome: %s, win_amount: %v, win_prob: %v}\n", rngResp.PrefOutcome, rngResp.WinAmount, rngResp.WinProb)

    // Convert RNG response to game-specific format and apply outcome
    gameRNGResp := RNGResponse{
        PrefOutcome: rngResp.PrefOutcome,
        WinAmount:   rngResp.WinAmount,
        WinProb:     rngResp.WinProb,
    }
    
    // Apply RNG outcome
    gs.ApplyRNGOutcome(gameRNGResp, rowReq.Action)

    maxPayout := rowReq.BetAmount * 10000
    if gs.AmountWon > maxPayout {
        gs.AmountWon = maxPayout
        log.Printf("Max payout reached: %v", maxPayout)
    }

    // Convert the internal reel-based game state to row-based for response
    rowBasedResponse := RowBasedSpinResponse{
        Status:  200,
        Message: "Success",
        Data:    gs.ConvertToRowBased(),
    }
    
    // Return row-based response
    return c.JSON(rowBasedResponse)
}



// // Move the SpinHandler to a method on RouteGroup
// func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
//     var req SpinRequest
//     if err := c.BodyParser(&req); err != nil {
//         log.Printf("Error parsing request: %v", err)
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//             "status":  400,
//             "message": "Invalid request format",
//         })
//     }

//     if req.BetAmount < 0.5 || (req.Action != "SPIN" && req.Action != "TRANSFORM") {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//             "status":  400,
//             "message": "Invalid bet amount or action",
//         })
//     }

//     // Use the shared settings service
//     rtp, err := rg.Settings.GetRTP(req.ClientID, req.Game.ID, req.PlayerID)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "status":  500,
//             "message": "Failed to fetch RTP",
//         })
//     }

//     fmt.Printf("RTP: %v\n", rtp)

//     gs := NewGameState(req.BetAmount, req.Game.Mode)
//     // Set game state fields from the request for both SPIN and TRANSFORM actions
//     gs.Cards = req.Cards
//     gs.FreeSpins = req.FreeSpins

//     // Only set ComboMultiplier for TRANSFORM actions
//     if req.Action == "TRANSFORM" {
//         gs.ComboMultiplier = req.ComboMultiplier
//     }

//     // Generate grid and calculate potential wins inside Spin/Transform
//     var potentialWins float64
//     if req.Action == "SPIN" {
//         potentialWins = gs.Spin()
//     } else if req.Action == "TRANSFORM" {
//         potentialWins = gs.Transform()
//     }

//     payoutMultiplier := potentialWins / req.BetAmount
//     fmt.Printf("Potential wins: %v, Payout multiplier: %v\n", potentialWins, payoutMultiplier)

//     // Use the shared RNG service
//     rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.Game.ID, req.PlayerID, rtp, payoutMultiplier, req.BetAmount)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "status":  500,
//             "message": "Failed to fetch RNG outcome",
//         })
//     }

//     fmt.Printf("RNG response: {pref_outcome: %s, win_amount: %v, win_prob: %v}\n", rngResp.PrefOutcome, rngResp.WinAmount, rngResp.WinProb)

//     // Convert RNG response to game-specific format and apply outcome
//     gameRNGResp := RNGResponse{
//         PrefOutcome: rngResp.PrefOutcome,
//         WinAmount:   rngResp.WinAmount,
//         WinProb:     rngResp.WinProb,
//     }
    
//     // Apply RNG outcome
//     gs.ApplyRNGOutcome(gameRNGResp, req.Action)

//     maxPayout := req.BetAmount * 10000
//     if gs.AmountWon > maxPayout {
//         gs.AmountWon = maxPayout
//         log.Printf("Max payout reached: %v", maxPayout)
//     }

//     return c.JSON(SpinResponse{
//         Status:  200,
//         Message: "Success",
//         Data:    gs,
//     })
// }