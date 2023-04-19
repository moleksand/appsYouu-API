package service

import (
	"context"

	"github.com/steebchen/keskin-api/gqlgen"
	"github.com/steebchen/keskin-api/lib/sessctx"
	"github.com/steebchen/keskin-api/prisma"
)

func (r *Service) Reviews(ctx context.Context, obj *prisma.Service) (*gqlgen.ServiceReviewConnection, error) {
	viewer, err := sessctx.User(ctx)

	nodes := []*prisma.ServiceReview{}

	if obj.Deleted || err != nil {
		return &gqlgen.ServiceReviewConnection{
			Nodes: nodes,
		}, nil
	}

	viewerCompany := sessctx.Company(ctx)

	where := &prisma.ReviewWhereInput{
		Service: &prisma.ServiceWhereInput{
			ID: &obj.ID,
		},
	}

	if viewer.Type != prisma.UserTypeAdministrator {
		where.Or = []prisma.ReviewWhereInput{{
			Customer: &prisma.UserWhereInput{
				ID: &viewer.ID,
			},
		}}

		if viewer.Type == prisma.UserTypeManager {
			where.Or = append(where.Or, prisma.ReviewWhereInput{
				Customer: &prisma.UserWhereInput{
					Company: &prisma.CompanyWhereInput{
						ID: &viewerCompany,
					},
				},
			})
		} else if viewer.Type != prisma.UserTypeAdministrator {
			allowSharing := true
			approved := prisma.ReviewStatusApproved
			where.Or = append(where.Or, prisma.ReviewWhereInput{
				Customer: &prisma.UserWhereInput{
					AllowReviewSharing: &allowSharing,
				},
				Status: &approved,
			})
		}
	}

	reviews, err := r.Prisma.Reviews(&prisma.ReviewsParams{
		Where: where,
	}).Exec(ctx)

	if err != nil {
		return &gqlgen.ServiceReviewConnection{
			Nodes: nodes,
		}, err
	}

	for _, review := range reviews {
		clone := review
		nodes = append(nodes, clone.Convert().(*prisma.ServiceReview))
	}

	return &gqlgen.ServiceReviewConnection{
		Nodes: nodes,
	}, err
}
