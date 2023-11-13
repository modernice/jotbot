package mockgenerate

import (
	"sync"

	generate "github.com/modernice/jotbot/generate"
)

// MockMinifier represents a test double for mimicking the behavior of a
// minification service. It provides mechanisms to simulate and track
// invocations of the Minify method, allowing for controlled testing scenarios.
// The type offers ways to configure default responses or specific behaviors for
// subsequent calls, as well as inspecting the history of calls made to
// facilitate assertions and verifications in test cases. It can be initialized
// with default behavior, strict behavior that panics on unexpected calls, or by
// replicating the behavior of another [generate.Minifier] instance.
type MockMinifier struct {
	MinifyFunc *MinifierMinifyFunc
}

// NewMockMinifier creates a new instance of MockMinifier. It provides a default
// implementation for the Minify method that simply returns the input
// unmodified. This mock is intended to be used in tests where the actual
// minification behavior is not important but a mock implementation of the
// generate.Minifier interface is required. The behavior of the Minify method
// can be customized by setting hooks or default return values.
func NewMockMinifier() *MockMinifier {
	return &MockMinifier{
		MinifyFunc: &MinifierMinifyFunc{
			defaultHook: func([]byte) (r0 []byte, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockMinifier creates a new instance of a mock minifier that panics
// when its Minify method is called without an explicit expectation being set.
// It is useful in testing scenarios where the absence of an expected call to
// Minify should be immediately visible and result in a test failure. The
// returned mock is of type [*MockMinifier].
func NewStrictMockMinifier() *MockMinifier {
	return &MockMinifier{
		MinifyFunc: &MinifierMinifyFunc{
			defaultHook: func([]byte) ([]byte, error) {
				panic("unexpected invocation of MockMinifier.Minify")
			},
		},
	}
}

// NewMockMinifierFrom creates a new mock of the Minifier interface using an
// existing Minifier's Minify method as the default behavior for the mock's
// Minify function. It returns a pointer to the newly created MockMinifier. This
// is useful for testing, allowing one to create a mock based on the behavior of
// a real implementation, which can then be customized further if needed.
func NewMockMinifierFrom(i generate.Minifier) *MockMinifier {
	return &MockMinifier{
		MinifyFunc: &MinifierMinifyFunc{
			defaultHook: i.Minify,
		},
	}
}

// MinifierMinifyFunc represents the functionality to compress or transform a
// slice of bytes into a potentially smaller or optimized slice, along with an
// error that may occur during the process. It provides mechanisms to set and
// queue custom behaviors for minification, track historical calls, and define
// default responses.
type MinifierMinifyFunc struct {
	defaultHook func([]byte) ([]byte, error)
	hooks       []func([]byte) ([]byte, error)
	history     []MinifierMinifyFuncCall
	mutex       sync.Mutex
}

// Minify reduces the size of the input byte slice and returns the compacted
// version along with any error encountered during the process. It simulates the
// behavior of a minification process for testing purposes, allowing hook
// functions to be set for custom responses. It also records each call made to
// it, which can be retrieved later.
func (m *MockMinifier) Minify(v0 []byte) ([]byte, error) {
	r0, r1 := m.MinifyFunc.nextHook()(v0)
	m.MinifyFunc.appendCall(MinifierMinifyFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook assigns a new default hook function to be used for
// minification when no other hooks are in the queue.
func (f *MinifierMinifyFunc) SetDefaultHook(hook func([]byte) ([]byte, error)) {
	f.defaultHook = hook
}

// PushHook adds a new hook to the MinifierMinifyFunc's queue of hooks to be
// invoked sequentially during the minification process. Each hook is a function
// that takes a byte slice and returns a processed byte slice and an error if
// any. Once pushed, the hook will be used in the order it was added when Minify
// is called, before reverting to the default hook once all custom hooks have
// been consumed.
func (f *MinifierMinifyFunc) PushHook(hook func([]byte) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn establishes the return values for the default hook of a
// MinifierMinifyFunc, which will be used when no other hooks are present. This
// allows for specifying what the Minify method should return by default.
func (f *MinifierMinifyFunc) SetDefaultReturn(r0 []byte, r1 error) {
	f.SetDefaultHook(func([]byte) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn queues a return value pair to be used by the next invocation of
// Minify. The specified byte slice and error will be returned by Minify when it
// is called, bypassing the usual processing. This function is typically used in
// testing scenarios where deterministic output from the Minify method is
// desired. Subsequent calls to Minify will use further queued return values, or
// fallback to the default behavior if no more are queued.
func (f *MinifierMinifyFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func([]byte) ([]byte, error) {
		return r0, r1
	})
}

func (f *MinifierMinifyFunc) nextHook() func([]byte) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MinifierMinifyFunc) appendCall(r0 MinifierMinifyFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a slice of all invocations of the Minify function, including
// their arguments and results. Each invocation is represented as a
// MinifierMinifyFuncCall struct. The returned slice reflects the order in which
// the Minify function was called. This method can be useful for testing and
// debugging purposes to verify the sequence and outcomes of Minify calls.
func (f *MinifierMinifyFunc) History() []MinifierMinifyFuncCall {
	f.mutex.Lock()
	history := make([]MinifierMinifyFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MinifierMinifyFuncCall represents a record of a single invocation of a
// minification process, including the input data, the resulting minified data,
// and any error that occurred during the process. It provides methods to
// retrieve both the input arguments and output results of the minification
// call. This type is typically used for testing and debugging purposes to
// ensure that the minification function behaves as expected across various
// cases.
type MinifierMinifyFuncCall struct {
	Arg0    []byte
	Result0 []byte
	Result1 error
}

// Args returns the argument passed to the MinifierMinifyFuncCall, packaged in a
// slice of interface{} types. This argument represents the input data that was
// provided to the Minify function of a Minifier.
func (c MinifierMinifyFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns the slice of interface{} that contains the result of a Minify
// operation. The first element of the slice is the minified data ([]byte), and
// the second element is an error, if any occurred during the Minify operation.
func (c MinifierMinifyFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// MockService provides a mock implementation of the [generate.Service]
// interface for testing purposes. It allows for setting up expectations, which
// can specify the behavior of the `GenerateDoc` method when invoked with
// certain arguments. This includes setting default return values or specific
// behaviors for subsequent calls. It also records the invocation history of the
// `GenerateDoc` method, allowing you to inspect the calls made during a test to
// ensure correct interactions with the service.
type MockService struct {
	GenerateDocFunc *ServiceGenerateDocFunc
}

// NewMockService constructs a new mock implementation of the Service interface
// for testing purposes. It returns a [*MockService] with a default no-op hook
// for the GenerateDoc method, which can be customized with PushHook or
// SetDefaultHook for specific test cases.
func NewMockService() *MockService {
	return &MockService{
		GenerateDocFunc: &ServiceGenerateDocFunc{
			defaultHook: func(generate.Context) (r0 string, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockService creates and returns a new instance of MockService with a
// strict default behavior for its GenerateDoc method. This strict behavior
// causes a panic if the method is invoked without an explicit expectation set,
// indicating an unexpected call during testing.
func NewStrictMockService() *MockService {
	return &MockService{
		GenerateDocFunc: &ServiceGenerateDocFunc{
			defaultHook: func(generate.Context) (string, error) {
				panic("unexpected invocation of MockService.GenerateDoc")
			},
		},
	}
}

// NewMockServiceFrom creates a new instance of MockService by wrapping the
// provided generate.Service, allowing the underlying service's GenerateDoc
// method to be used as the default behavior for the mock's GenerateDoc method.
// This is typically used in testing scenarios where you want to create a mock
// that starts with predefined behavior from an existing service implementation.
// It returns a pointer to the newly created MockService.
func NewMockServiceFrom(i generate.Service) *MockService {
	return &MockService{
		GenerateDocFunc: &ServiceGenerateDocFunc{
			defaultHook: i.GenerateDoc,
		},
	}
}

// ServiceGenerateDocFunc encapsulates a behavior to generate a document based
// on the provided context, returning the document as a string along with any
// error that occurred during generation. It supports setting a default
// behavior, pushing custom behaviors to be invoked in sequence, and maintaining
// a history of its invocations and outcomes.
type ServiceGenerateDocFunc struct {
	defaultHook func(generate.Context) (string, error)
	hooks       []func(generate.Context) (string, error)
	history     []ServiceGenerateDocFuncCall
	mutex       sync.Mutex
}

// GenerateDoc invokes the configured hook for generating a document within a
// given context and returns the generated document along with any error that
// occurred during generation. It records each invocation to allow for later
// inspection of the call history.
func (m *MockService) GenerateDoc(v0 generate.Context) (string, error) {
	r0, r1 := m.GenerateDocFunc.nextHook()(v0)
	m.GenerateDocFunc.appendCall(ServiceGenerateDocFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook establishes the fallback behavior for generating documentation
// within the service when no other hooks have been defined. It accepts a
// function that takes a [generate.Context] and returns a document identifier as
// a string and an error, if any. This function will be invoked by default
// during the document generation process unless overridden by subsequent hooks.
func (f *ServiceGenerateDocFunc) SetDefaultHook(hook func(generate.Context) (string, error)) {
	f.defaultHook = hook
}

// PushHook appends a new hook function to the ServiceGenerateDocFunc's hook
// queue. Each hook takes a [generate.Context] and returns a string and an
// error. When the GenerateDoc method of MockService is called, hooks are
// invoked in the order they were added, with each subsequent call using the
// next hook in the queue. If no hooks are present, the default hook is used.
func (f *ServiceGenerateDocFunc) PushHook(hook func(generate.Context) (string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn sets the default return values for the GenerateDoc function
// of the service mock. The specified values will be returned whenever the
// function is invoked without a specific hook being configured. This method is
// useful for setting up expected behavior in a controlled testing environment.
func (f *ServiceGenerateDocFunc) SetDefaultReturn(r0 string, r1 error) {
	f.SetDefaultHook(func(generate.Context) (string, error) {
		return r0, r1
	})
}

// PushReturn queues a return value to be provided on the next invocation of the
// GenerateDoc method of the mock service. The specified return values will be
// used once, and in the order they were added, before falling back to the
// default behavior or subsequent hooks if they exist. The function ensures
// thread-safety by managing concurrent access to the hook queue.
func (f *ServiceGenerateDocFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(generate.Context) (string, error) {
		return r0, r1
	})
}

func (f *ServiceGenerateDocFunc) nextHook() func(generate.Context) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ServiceGenerateDocFunc) appendCall(r0 ServiceGenerateDocFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a slice of all invocations of the GenerateDoc method,
// including their input contexts and output results. Each entry in the returned
// slice is a record of type [ServiceGenerateDocFuncCall], which captures the
// arguments and return values of a single call.
func (f *ServiceGenerateDocFunc) History() []ServiceGenerateDocFuncCall {
	f.mutex.Lock()
	history := make([]ServiceGenerateDocFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ServiceGenerateDocFuncCall represents a record of an invocation to generate a
// document within a given context, including the arguments provided, and the
// results produced. It captures the input context and the output in the form of
// a document identifier and any error that might have occurred during the
// document generation process. This type is typically used for auditing or
// debugging purposes to track the history of function calls and their outcomes.
type ServiceGenerateDocFuncCall struct {
	Arg0    generate.Context
	Result0 string
	Result1 error
}

// Args returns the input arguments of the ServiceGenerateDoc function call as a
// slice of interface{} types, where each element in the slice corresponds to an
// individual argument.
func (c ServiceGenerateDocFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns the outcomes of a ServiceGenerateDocFuncCall as a slice of
// empty interface types. This includes the generated document identifier and
// any error encountered during the document generation process, encapsulated in
// a generic way to allow for flexibility in handling the results.
func (c ServiceGenerateDocFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
