package main

import (
	"context"
	"encoding/base64"
	"log"

	"gopkg.in/gomail.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Email struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EmailSpec   `json:"spec"`
	Status            EmailStatus `json:"status,omitempty"`
}

type EmailSpec struct {
	SenderConfigRef string `json:"senderConfigRef"`
	RecipientEmail  string `json:"recipientEmail"`
	Subject         string `json:"subject"`
	Body            string `json:"body"`
}

type EmailStatus struct {
	DeliveryStatus string `json:"deliveryStatus"`
	MessageId      string `json:"messageId"`
	Error          string `json:"error,omitempty"`
}

type EmailSenderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EmailSenderConfigSpec `json:"spec"`
}

type EmailSenderConfigSpec struct {
	ApiTokenSecretRef string `json:"apiTokenSecretRef"`
	SenderEmail       string `json:"senderEmail"`
}

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Failed to build Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes dynamic client: %v", err)
	}

	namespace := "default"
	emailResourceName := "emailsender-01.com"
	senderResourceName := "yemi1842@gmail.com"

	emailConfig, err := getEmailConfig(dynamicClient, namespace, emailResourceName)
	if err != nil {
		log.Fatalf("Error getting EmailConfig: %v", err)
	}

	senderConfig, err := getSenderConfig(dynamicClient, namespace, senderResourceName)
	if err != nil {
		log.Fatalf("Error getting SenderConfig: %v", err)
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), senderConfig.Spec.ApiTokenSecretRef, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting secret: %v", err)
	}

	apiToken, err := base64.StdEncoding.DecodeString(string(secret.Data["apiToken"]))
	if err != nil {
		log.Fatalf("Failed to decode API token: %v", err)
	}

	if err := sendEmail(senderConfig.Spec.SenderEmail, emailConfig.Spec.RecipientEmail, emailConfig.Spec.Subject, emailConfig.Spec.Body, string(apiToken)); err != nil {
		log.Printf("Failed to send email: %v", err)
		updateEmailStatus(dynamicClient, namespace, emailResourceName, "Failed", "", err.Error())
		return
	}

	log.Println("Email sent successfully")
	updateEmailStatus(dynamicClient, namespace, emailResourceName, "Successful", "17", "")
}

func getEmailConfig(dynamicClient dynamic.Interface, namespace, name string) (*Email, error) {
	emailResource := schema.GroupVersionResource{Group: "mailerlite.task.com", Version: "v1", Resource: "emails"}
	unstruct, err := dynamicClient.Resource(emailResource).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	email := &Email{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(), email)
	return email, err
}

func getSenderConfig(dynamicClient dynamic.Interface, namespace, name string) (*EmailSenderConfig, error) {
	senderResource := schema.GroupVersionResource{Group: "mailerlite.task.com", Version: "v1", Resource: "emailsenderconfigs"}
	unstruct, err := dynamicClient.Resource(senderResource).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	sender := &EmailSenderConfig{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(), sender)
	return sender, err
}

func sendEmail(from string, to string, subject string, body string, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.mailersend.com", 587, from, token)

	return d.DialAndSend(m)
}

func updateEmailStatus(dynamicClient dynamic.Interface, namespace, name, status, messageId, error string) {
	emailResource := schema.GroupVersionResource{Group: "mailerlite.task.com", Version: "v1", Resource: "emails"}
	email, err := getEmailConfig(dynamicClient, namespace, name)
	if err != nil {
		log.Printf("Error getting EmailConfig for update: %v", err)
		return
	}

	email.Status.DeliveryStatus = status
	email.Status.MessageId = messageId
	email.Status.Error = error

	unstruct, err := runtime.DefaultUnstructuredConverter.ToUnstructured(email)
	if err != nil {
		log.Printf("Error converting EmailConfig to unstructured: %v", err)
		return
	}

	_, err = dynamicClient.Resource(emailResource).Namespace(namespace).UpdateStatus(context.TODO(), &unstructured.Unstructured{Object: unstruct}, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error updating EmailConfig status: %v", err)
	}
}
