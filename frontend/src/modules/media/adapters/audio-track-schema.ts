import { z } from 'zod';
import {
  audioChunkBytes,
  audioTrackKinds,
  audioTrackStates,
  audioUploadPath,
  maximumAudioBytes,
  validAudioLabel,
  validAudioLanguage,
} from '../domain/audio-track';
import { validMediaId } from '../domain/video';
import { isoTimestampSchema } from '../../../adapters/api/timestamp-schema';

const mediaIdSchema = z.uuid().refine(validMediaId);

export const audioTrackSchema = z
  .object({
    trackId: mediaIdSchema,
    videoId: mediaIdSchema,
    kind: z.enum(audioTrackKinds),
    language: z.string().refine(validAudioLanguage),
    label: z.string().refine(validAudioLabel),
    isDefault: z.boolean(),
    state: z.enum(audioTrackStates),
    updatedAt: isoTimestampSchema,
  })
  .strict();

export const audioTracksSchema = z.object({ items: z.array(audioTrackSchema) }).strict();

export const audioUploadTicketSchema = z
  .object({
    trackId: mediaIdSchema,
    uploadUrl: z.string(),
    chunkBytes: z.literal(audioChunkBytes),
    maxBytes: z.literal(maximumAudioBytes),
    expiresAt: isoTimestampSchema,
  })
  .strict()
  .refine((ticket) => ticket.uploadUrl === audioUploadPath(ticket.trackId));
