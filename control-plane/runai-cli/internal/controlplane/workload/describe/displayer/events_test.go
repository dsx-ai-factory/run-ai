package displayer_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("EventBuilder", func() {
	It("should build an Event correctly", func() {
		createdAt := time.Now()
		event := displayer.NewEventBuilder().
			CreatedAt(createdAt).
			Id("event1").
			Message("Test event").
			Reason("Testing").
			Source("TestSource").
			InvolvedObject(displayer.InvolvedObject{Uid: "obj1", Kind: "Pod", Name: "pod1", Namespace: "default"}).
			Build()

		Expect(event.CreatedAt).To(Equal(createdAt))
		Expect(*event.Id).To(Equal("event1"))
		Expect(event.Message).To(Equal("Test event"))
		Expect(*event.Reason).To(Equal("Testing"))
		Expect(*event.Source).To(Equal("TestSource"))
		Expect(event.InvolvedObject.Uid).To(Equal("obj1"))
		Expect(event.InvolvedObject.Kind).To(Equal("Pod"))
		Expect(event.InvolvedObject.Name).To(Equal("pod1"))
		Expect(event.InvolvedObject.Namespace).To(Equal("default"))
	})
})

var _ = Describe("InvolvedObjectBuilder", func() {
	It("should build an InvolvedObject correctly", func() {
		obj := displayer.NewInvolvedObjectBuilder().
			Uid("obj1").
			Kind("Pod").
			Name("pod1").
			Namespace("default").
			Build()

		Expect(obj.Uid).To(Equal("obj1"))
		Expect(obj.Kind).To(Equal("Pod"))
		Expect(obj.Name).To(Equal("pod1"))
		Expect(obj.Namespace).To(Equal("default"))
	})
})

var _ = Describe("EventsBuilder", func() {
	It("should build a collection of events", func() {
		event1 := displayer.NewEventBuilder().Message("event1").Build()
		event2 := displayer.NewEventBuilder().Message("event2").Build()

		events := displayer.NewEventsBuilder().
			Add(event1).
			Add(event2).
			Build()

		Expect(events).To(HaveLen(2))
		Expect(events[0].Message).To(Equal("event1"))
		Expect(events[1].Message).To(Equal("event2"))
	})

	It("should set entire collection with Set method", func() {
		event1 := displayer.NewEventBuilder().Message("event1").Build()
		event2 := displayer.NewEventBuilder().Message("event2").Build()

		events := displayer.NewEventsBuilder().
			Set([]displayer.Event{event1}).
			Add(event2).
			Build()

		Expect(events).To(HaveLen(2))
	})
})
