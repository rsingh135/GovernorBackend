-- Migration 006: Add freeform purchase-guideline policy text.

ALTER TABLE policies
ADD COLUMN IF NOT EXISTS purchase_guideline TEXT NOT NULL DEFAULT '';
