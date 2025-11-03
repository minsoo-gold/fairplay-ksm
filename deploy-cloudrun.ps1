# windows용 real server build
# deploy-cloudrun.ps1
param(
    [string]$ProjectID = "fairplaystreaming",
    [string]$Region = "asia-northeast3",
    [string]$Service = "ksm-server"
    # [string]$Firestore_db = "multi-drm-server"
)

gcloud config set project $ProjectID

gcloud run deploy $Service `
    --source . `
    --project $ProjectID `
    --region $Region `
    --allow-unauthenticated `
    --cpu 1 `
    --memory 512Mi `
    --min-instances 0 `
    --max-instances 50 `
    --execution-environment gen2 `
    --set-env-vars "GOOGLE_CLOUD_PROJECT=$ProjectID,FIRESTORE_DB=default" `
    --set-env-vars "REGION=$Region" `
    --ingress all

$Url = gcloud run services describe $Service `
    --project $ProjectID `
    --region $Region `
    --format='value(status.url)'

Write-Host "Deployment completed! Access your service at: $Url"
