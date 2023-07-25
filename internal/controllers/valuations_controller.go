package controllers

import "github.com/gofiber/fiber/v2"

type ValuationsController struct {
}

func (c *ValuationsController) GetValuations(ctx *fiber.Ctx) error {

}

func (c *ValuationsController) GetOffers(ctx *fiber.Ctx) error {

}

func NewValuationsController() *ValuationsController {
	return &ValuationsController{}
}
