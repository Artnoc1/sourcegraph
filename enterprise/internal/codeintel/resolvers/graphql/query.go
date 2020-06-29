package graphql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

// DefaultReferencesPageSize is the reference result page size when no limit is supplied.
const DefaultReferencesPageSize = 100

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

// QueryResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type QueryResolver struct {
	resolver         resolvers.QueryResolver
	locationResolver *CachedLocationResolver
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func NewQueryResolver(resolver resolvers.QueryResolver, locationResolver *CachedLocationResolver) gql.GitBlobLSIFDataResolver {
	return &QueryResolver{
		resolver:         resolver,
		locationResolver: locationResolver,
	}
}

func (r *QueryResolver) ToGitTreeLSIFData() (gql.GitTreeLSIFDataResolver, bool) { return r, true }
func (r *QueryResolver) ToGitBlobLSIFData() (gql.GitBlobLSIFDataResolver, bool) { return r, true }

//
// START EXPERIMENT
//

func (r *QueryResolver) NavView(ctx context.Context, args gql.LSIFNavViewArgs) (gql.NavViewConnectionResolver, error) {
	rangeViews, err := r.resolver.NavView(ctx)
	if err != nil {
		return nil, err
	}

	return &navViewConnectionResolver{
		rangeViews:       rangeViews,
		locationResolver: r.locationResolver,
	}, nil
}

type navViewConnectionResolver struct {
	rangeViews       []resolvers.RangeView
	locationResolver *CachedLocationResolver
}

func (r *navViewConnectionResolver) Nodes(ctx context.Context) ([]gql.NavRangeResolver, error) {
	var resolvers []gql.NavRangeResolver
	for _, rv := range r.rangeViews {
		resolvers = append(resolvers, &navRangeResolver{
			rangeView:        rv,
			locationResolver: r.locationResolver,
		})
	}

	return resolvers, nil
}

type navRangeResolver struct {
	rangeView        resolvers.RangeView
	locationResolver *CachedLocationResolver
}

func (r *navRangeResolver) Range(ctx context.Context) (gql.RangeResolver, error) {
	return gql.NewRangeResolver(convertRange(r.rangeView.Range)), nil
}

func (r *navRangeResolver) Definitions(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.rangeView.Definitions, nil, r.locationResolver), nil
}

func (r *navRangeResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (r *navRangeResolver) Hover(ctx context.Context) (gql.HoverResolver, error) {
	return NewHoverResolver(r.rangeView.HoverText, convertRange(r.rangeView.HoverRange)), nil
}

//
// END EXPERIMENT
//

func (r *QueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, err := r.resolver.Definitions(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	return NewLocationConnectionResolver(locations, nil, r.locationResolver), nil
}

func (r *QueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}
	cursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.resolver.References(ctx, int(args.Line), int(args.Character), limit, cursor)
	if err != nil {
		return nil, err
	}

	return NewLocationConnectionResolver(locations, strPtr(cursor), r.locationResolver), nil
}

func (r *QueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.HoverResolver, error) {
	text, rx, exists, err := r.resolver.Hover(ctx, int(args.Line), int(args.Character))
	if err != nil || !exists {
		return nil, err
	}

	return NewHoverResolver(text, convertRange(rx)), nil
}

func (r *QueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) (gql.DiagnosticConnectionResolver, error) {
	limit := derefInt32(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	diagnostics, totalCount, err := r.resolver.Diagnostics(ctx, limit)
	if err != nil {
		return nil, err
	}

	return NewDiagnosticConnectionResolver(diagnostics, totalCount, r.locationResolver), nil
}
