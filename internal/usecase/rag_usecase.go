package usecase

import (
	"backend/internal/repository"
	"context"
	"fmt"
	"strings"
)

type RAGUsecase struct {
	tripRepo repository.TripRepository
}

func NewRAGUsecase(
	tripRepo repository.TripRepository,
) *RAGUsecase {
	return &RAGUsecase{
		tripRepo: tripRepo,
	}
}

func (u *RAGUsecase) BuildContext(
	ctx context.Context,
	destination string,
) (string, error) {

	trips, err := u.tripRepo.SearchSimilarTrips(
		ctx,
		destination,
	)

	if err != nil {
		return "", err
	}

	var builder strings.Builder

	for _, trip := range trips {

		builder.WriteString(
			fmt.Sprintf(
				"Trip: %s -> %s\n",
				trip.From,
				trip.To,
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"Duration: %d Days\n",
				trip.Duration,
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"Trip Type: %s\n",
				trip.TripType,
			),
		)

		for _, plan := range trip.Plans {

			builder.WriteString(
				fmt.Sprintf(
					"Day %d: %s\n",
					plan.DayNumber,
					plan.Title,
				),
			)

			builder.WriteString(
				fmt.Sprintf(
					"%s\n",
					plan.Description,
				),
			)
		}

		builder.WriteString("\n-------------------\n")
	}

	return builder.String(), nil
}