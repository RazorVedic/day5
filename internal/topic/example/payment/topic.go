package payment

// Topic represents a pubsub topic (replacement for pubsub.Topic)
type Topic struct {
	Name string
}

// PublishTopics keeps a list of pubsub topics which main publishes to
var PublishTopics = []string{
	Created.Name,
	Success.Name,
	Failed.Name,
}

// Created for the payment created event
var Created = Topic{
	Name: "payment_created",
}

// Success for the payment success event
var Success = Topic{
	Name: "payment_success",
}

// Failed for the payment failed event
var Failed = Topic{
	Name: "payment_failed",
}
