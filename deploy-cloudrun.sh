#!/usr/bin/env bash
set -euo pipefail

PROJECT_ID="fairplaystreaming"
REGION="asia-northeast3"          # 서울
SERVICE="ksm-server"

# 빌드 & 배포(소스에서 빌드)
gcloud run deploy "$SERVICE" \
  --source . \
  --project "$PROJECT_ID" \
  --region "$REGION" \
  --allow-unauthenticated \
  --cpu=1 \
  --memory=512Mi \
  --min-instances=0 \
  --max-instances=50 \
  --execution-environment gen2 \
  --set-env-vars "GOOGLE_CLOUD_PROJECT=$PROJECT_ID,FIRESTORE_DB=default" \
  --set-env-vars "REGION=$REGION" \
  --ingress all

# URL 확인
gcloud run services describe "$SERVICE" --project "$PROJECT_ID" --region "$REGION" --format='value(status.url)'