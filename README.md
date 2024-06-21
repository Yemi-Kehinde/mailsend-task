# mailsend-task

Email Operator for Kubernetes
Overview
The Email Operator automates email sending via transactional email providers like MailerSend in Kubernetes clusters. It manages email sender configurations and sends emails based on defined resources.

Features

- Manages EmailSenderConfig and Email custom resources.
- Sends emails using MailerSend API.
- Updates Email resource status based on delivery status.

Deployment
Prerequisites
- Kubernetes cluster
- kubectl configured to connect to your cluster

Step-by-Step Deployment
Deploy CRDs:


kubectl apply -f emailsenderconfig_crd.yaml
kubectl apply -f email_crd.yaml


Deploy Email Operator:

kubectl apply -f operator-deployment.yaml

Verify Deployment:

kubectl get pods -l app=email-operator

Usage
Configuration
1. Create EmailSenderConfig Resource:


apiVersion: mailerlite.task.com/v1
kind: EmailSenderConfig
metadata:
  name: emailsender-01.com
spec:
  senderEmail: yemi1842@gmail.com
  apiTokenSecretRef: mailersend-secret


2. Send Email Using Email Resource:

apiVersion: mailerlite.task.com/v1
kind: Email
metadata:
  name: email-01
spec:
  senderConfigRef: emailsender-01.com
  recipientEmail: okreceive@gmail.co
  subject: "hello"
  body: "Hello this is email body"


Check email status:
kubectl get email example-email -o yaml

Testing
- Create instances of EmailSenderConfig and Email to test functionality.
- Verify email sending and status updates.

Cleanup
Remove Email Operator and resources:

- kubectl delete -f operator-deployment.yaml
- kubectl delete -f emailsenderconfig_crd.yaml
- kubectl delete -f email_crd.yaml

This README.md provides a quick guide to deploying, configuring, using, testing, and cleaning up the Email Operator in your Kubernetes cluster.