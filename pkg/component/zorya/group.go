package zorya

// Group is a collection of routes that share a common prefix and set of
// operation modifiers, middlewares, and transformers.
//
// This is useful for grouping related routes together and applying common
// settings to them. For example, you might create a group for all routes that
// require authentication.
type Group struct {
	API
	prefixes     []string
	adapter      Adapter
	modifiers    []func(o *BaseRoute, next func(*BaseRoute))
	middlewares  Middlewares
	transformers []Transformer
}

// Transformer is a function that transforms response bodies before serialization.
// Transformers are run in the order they are added.
type Transformer func(ctx Context, status string, v any) (any, error)

// NewGroup creates a new group of routes with the given prefixes, if any. A
// group enables a collection of operations to have the same prefix and share
// operation modifiers, middlewares, and transformers.
//
//	grp := zorya.NewGroup(api, "/v1")
//	grp.UseMiddleware(authMiddleware)
//
//	zorya.Get(grp, "/users", func(ctx context.Context, input *MyInput) (*MyOutput, error) {
//		// Your code here...
//	})
func NewGroup(api API, prefixes ...string) *Group {
	group := &Group{API: api, prefixes: prefixes}
	group.adapter = &groupAdapter{Adapter: api.Adapter(), group: group}
	if len(prefixes) > 0 {
		group.UseModifier(PrefixModifier(prefixes))
	}

	return group
}

func (g *Group) Adapter() Adapter {
	return g.adapter
}

// PrefixModifier provides a fan-out to register one or more operations with
// the given prefix for every one operation added to a group.
func PrefixModifier(prefixes []string) func(o *BaseRoute, next func(*BaseRoute)) {
	return func(o *BaseRoute, next func(*BaseRoute)) {
		for _, prefix := range prefixes {
			modified := *o
			modified.Path = prefix + modified.Path
			next(&modified)
		}
	}
}

// groupAdapter is an Adapter wrapper that registers multiple operation handlers
// with the underlying adapter based on the group's prefixes.
type groupAdapter struct {
	Adapter
	group *Group
}

func (a *groupAdapter) Handle(route *BaseRoute, handler func(Context)) {
	a.group.ModifyOperation(route, func(route *BaseRoute) {
		a.Adapter.Handle(route, handler)
	})
}

// ModifyOperation runs all operation modifiers in the group on the given
// route, in the order they were added. This is useful for modifying a route
// before it is registered with the router.
func (g *Group) ModifyOperation(route *BaseRoute, next func(*BaseRoute)) {
	chain := func(route *BaseRoute) {
		// Call the final handler.
		next(route)
	}

	for i := len(g.modifiers) - 1; i >= 0; i-- {
		// Use an inline func to provide a closure around the index & chain.
		func(i int, n func(*BaseRoute)) {
			chain = func(route *BaseRoute) { g.modifiers[i](route, n) }
		}(i, chain)
	}

	chain(route)
}

// UseModifier adds an operation modifier function to the group that will be run
// on all operations in the group. Use this to modify the operation before it is
// registered with the router. This behaves similar to middleware in that you
// should invoke `next` to continue the chain. Skip it to prevent the operation
// from being registered, and call multiple times for a fan-out effect.
func (g *Group) UseModifier(modifier func(o *BaseRoute, next func(*BaseRoute))) {
	g.modifiers = append(g.modifiers, modifier)
}

// UseSimpleModifier adds an operation modifier function to the group that
// will be run on all operations in the group. Use this to modify the operation
// before it is registered with the router.
func (g *Group) UseSimpleModifier(modifier func(o *BaseRoute)) {
	g.modifiers = append(g.modifiers, func(o *BaseRoute, next func(*BaseRoute)) {
		modifier(o)
		next(o)
	})
}

// UseMiddleware adds one or more middleware functions to the group that will be
// run on all operations in the group. Use this to add common functionality to
// all operations in the group, e.g. authentication/authorization.
func (g *Group) UseMiddleware(middlewares ...func(ctx Context, next func(Context))) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) Middlewares() Middlewares {
	m := append(Middlewares{}, g.API.Middlewares()...)

	return append(m, g.middlewares...)
}

// UseTransformer adds one or more transformer functions to the group that will
// be run on all responses in the group.
func (g *Group) UseTransformer(transformers ...Transformer) {
	g.transformers = append(g.transformers, transformers...)
}

// Transform runs all transformers in the group on the response, in the order
// they were added, then chains to the parent API's transformers.
func (g *Group) Transform(ctx Context, status string, v any) (any, error) {
	// Run group-specific transformers first.
	for _, transformer := range g.transformers {
		var err error
		v, err = transformer(ctx, status, v)
		if err != nil {
			return v, err
		}
	}

	// Chain to parent API transformers.
	return g.API.Transform(ctx, status, v)
}
