-- PostgreSQL 16+ function to convert UUID to ULID string format
-- Based on the pULID implementation from pixie-sh/ulid-go
-- 
-- pULID Structure:
--   6 bytes: timestamp (epoch in milliseconds)
--   2 bytes: scope (reserved for entity types/namespaces)
--   8 bytes: entropy
--
-- This function assumes the input UUID follows the pULID format where:
-- - Bytes 0-5: timestamp
-- - Bytes 6-7: scope 
-- - Bytes 8-15: entropy

CREATE OR REPLACE FUNCTION uuid_to_ulid(input_uuid UUID)
    RETURNS TEXT
    LANGUAGE plpgsql
    IMMUTABLE
    STRICT
    SECURITY DEFINER
AS
$$
DECLARE
    -- Base32 encoding table used by ULID (Crockford's base32)
    encoding CONSTANT TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';

    -- Binary representation of the UUID
    binary_data       BYTEA;
    uuid_text         TEXT;

    -- Individual bytes for processing (using ARRAY for better performance)
    bytes             INT[16];

    -- Result ULID characters (using ARRAY for efficient concatenation)
    result_chars      TEXT[26];

BEGIN
    -- Validate input
    IF input_uuid IS NULL THEN
        RETURN NULL;
    END IF;

    -- Convert UUID to text and validate format
    uuid_text := input_uuid::TEXT;

    -- Validate UUID format (8-4-4-4-12 pattern with hyphens)
    IF NOT uuid_text ~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$' THEN
        RAISE EXCEPTION 'Invalid UUID format: %', uuid_text
            USING ERRCODE = '22023', -- invalid_parameter_value
                HINT = 'UUID must be in format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
    END IF;

    -- Convert UUID to binary (remove hyphens and decode hex)
    BEGIN
        binary_data := decode(replace(uuid_text, '-', ''), 'hex');
    EXCEPTION
        WHEN OTHERS THEN
            RAISE EXCEPTION 'Failed to decode UUID hex data: %', uuid_text
                USING ERRCODE = '22023', -- invalid_parameter_value
                    HINT = 'UUID contains invalid hexadecimal characters';
    END;

    -- Validate binary data length
    IF length(binary_data) != 16 THEN
        RAISE EXCEPTION 'Invalid UUID binary length: expected 16 bytes, got %', length(binary_data)
            USING ERRCODE = '22023'; -- invalid_parameter_value
    END IF;

    -- Extract individual bytes using array assignment for better performance
    FOR i IN 1..16
        LOOP
            bytes[i] := get_byte(binary_data, i - 1);
        END LOOP;

    -- Encode timestamp part (first 10 characters of ULID - bytes 0-5)
    -- Following the exact bit manipulation from the Go implementation
    result_chars[1] := substr(encoding, ((bytes[1] & 224) >> 5) + 1, 1); -- dst[0]
    result_chars[2] := substr(encoding, (bytes[1] & 31) + 1, 1); -- dst[1]
    result_chars[3] := substr(encoding, ((bytes[2] & 248) >> 3) + 1, 1); -- dst[2]
    result_chars[4] := substr(encoding, (((bytes[2] & 7) << 2) | ((bytes[3] & 192) >> 6)) + 1, 1); -- dst[3]
    result_chars[5] := substr(encoding, ((bytes[3] & 62) >> 1) + 1, 1); -- dst[4]
    result_chars[6] := substr(encoding, (((bytes[3] & 1) << 4) | ((bytes[4] & 240) >> 4)) + 1, 1); -- dst[5]
    result_chars[7] := substr(encoding, (((bytes[4] & 15) << 1) | ((bytes[5] & 128) >> 7)) + 1, 1); -- dst[6]
    result_chars[8] := substr(encoding, ((bytes[5] & 124) >> 2) + 1, 1); -- dst[7]
    result_chars[9] := substr(encoding, (((bytes[5] & 3) << 3) | ((bytes[6] & 224) >> 5)) + 1, 1); -- dst[8]
    result_chars[10] := substr(encoding, (bytes[6] & 31) + 1, 1);
    -- dst[9]

    -- Encode entropy part (last 16 characters of ULID - bytes 6-15)
    result_chars[11] := substr(encoding, ((bytes[7] & 248) >> 3) + 1, 1); -- dst[10]
    result_chars[12] := substr(encoding, (((bytes[7] & 7) << 2) | ((bytes[8] & 192) >> 6)) + 1, 1); -- dst[11]
    result_chars[13] := substr(encoding, ((bytes[8] & 62) >> 1) + 1, 1); -- dst[12]
    result_chars[14] := substr(encoding, (((bytes[8] & 1) << 4) | ((bytes[9] & 240) >> 4)) + 1, 1); -- dst[13]
    result_chars[15] := substr(encoding, (((bytes[9] & 15) << 1) | ((bytes[10] & 128) >> 7)) + 1, 1); -- dst[14]
    result_chars[16] := substr(encoding, ((bytes[10] & 124) >> 2) + 1, 1); -- dst[15]
    result_chars[17] := substr(encoding, (((bytes[10] & 3) << 3) | ((bytes[11] & 224) >> 5)) + 1, 1); -- dst[16]
    result_chars[18] := substr(encoding, (bytes[11] & 31) + 1, 1); -- dst[17]
    result_chars[19] := substr(encoding, ((bytes[12] & 248) >> 3) + 1, 1); -- dst[18]
    result_chars[20] := substr(encoding, (((bytes[12] & 7) << 2) | ((bytes[13] & 192) >> 6)) + 1, 1); -- dst[19]
    result_chars[21] := substr(encoding, ((bytes[13] & 62) >> 1) + 1, 1); -- dst[20]
    result_chars[22] := substr(encoding, (((bytes[13] & 1) << 4) | ((bytes[14] & 240) >> 4)) + 1, 1); -- dst[21]
    result_chars[23] := substr(encoding, (((bytes[14] & 15) << 1) | ((bytes[15] & 128) >> 7)) + 1, 1); -- dst[22]
    result_chars[24] := substr(encoding, ((bytes[15] & 124) >> 2) + 1, 1); -- dst[23]
    result_chars[25] := substr(encoding, (((bytes[15] & 3) << 3) | ((bytes[16] & 224) >> 5)) + 1, 1); -- dst[24]
    result_chars[26] := substr(encoding, (bytes[16] & 31) + 1, 1);
    -- dst[25]

    -- Convert array to string using array_to_string for better performance
    RETURN array_to_string(result_chars, '');

END;
$$;

-- Create a comment explaining the function
COMMENT ON FUNCTION uuid_to_ulid(UUID) IS
    'Converts a UUID that follows pULID format to ULID string representation. '
    'Based on pixie-sh/ulid-go implementation with Crockford Base32 encoding. '
    'Raises exceptions for invalid UUIDs or conversion errors.';

-- Function to convert ULID string back to UUID format
CREATE OR REPLACE FUNCTION ulid_to_uuid(input_ulid TEXT)
    RETURNS UUID
    LANGUAGE plpgsql
    IMMUTABLE
    STRICT
    SECURITY DEFINER
AS
$$
DECLARE
    -- Crockford Base32 decoding table (c2b32 from Go implementation)
    -- 0xFF indicates invalid character
    c2b32 CONSTANT INT[] := ARRAY [
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255, 255, 255, 255, 255, 255,
        255, 10, 11, 12, 13, 14, 15, 16, 17, 255, 18, 19, 255, 20, 21, 255,
        22, 23, 24, 25, 26, 255, 27, 28, 29, 30, 31, 255, 255, 255, 255, 255,
        255, 10, 11, 12, 13, 14, 15, 16, 17, 255, 18, 19, 255, 20, 21, 255,
        22, 23, 24, 25, 26, 255, 27, 28, 29, 30, 31, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
        255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255
        ];
    binary_data    BYTEA;
    v              INT[26]; -- Character values array
    result_bytes   INT[16]; -- Result bytes array
    hex_string     TEXT;
    i              INT;
    char_code      INT;

BEGIN
    -- Validate input
    IF input_ulid IS NULL THEN
        RETURN NULL;
    END IF;

    -- Validate ULID format and length
    IF length(input_ulid) != 26 THEN
        RAISE EXCEPTION 'Invalid ULID length: expected 26 characters, got %', length(input_ulid)
            USING ERRCODE = '22023', -- invalid_parameter_value
                HINT = 'ULID must be exactly 26 characters long';
    END IF;

    -- Validate ULID format (only valid Crockford Base32 characters)
    IF NOT input_ulid ~ '^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$' THEN
        RAISE EXCEPTION 'Invalid ULID format: contains invalid characters in %', input_ulid
            USING ERRCODE = '22023', -- invalid_parameter_value
                HINT = 'ULID must contain only characters: 0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    END IF;

    -- Check timestamp overflow (first character must be <= 7)
    IF ascii(substr(input_ulid, 1, 1)) > ascii('7') THEN
        RAISE EXCEPTION 'ULID timestamp overflow: first character ''%'' > 7', substr(input_ulid, 1, 1)
            USING ERRCODE = '22023', -- invalid_parameter_value
                HINT = 'ULID timestamp is too large';
    END IF;

    -- Convert each character to its corresponding value
    FOR i IN 1..26
        LOOP
            char_code := ascii(substr(input_ulid, i, 1)) + 1;
            -- +1 for 1-based array indexing

            -- Validate character and get its value
            IF char_code < 1 OR char_code > 256 OR c2b32[char_code] = 255 THEN
                RAISE EXCEPTION 'Invalid character ''%'' at position % in ULID', substr(input_ulid, i, 1), i
                    USING ERRCODE = '22023', -- invalid_parameter_value
                        HINT = 'Character is not valid in Crockford Base32 encoding';
            END IF;

            v[i] := c2b32[char_code];
        END LOOP;

    -- Decode timestamp (48 bits) - following Go implementation exactly with proper bit masking
    result_bytes[1] := ((v[1] << 5) | v[2]) & 255;
    result_bytes[2] := ((v[3] << 3) | (v[4] >> 2)) & 255;
    result_bytes[3] := ((v[4] << 6) | (v[5] << 1) | (v[6] >> 4)) & 255;
    result_bytes[4] := ((v[6] << 4) | (v[7] >> 1)) & 255;
    result_bytes[5] := ((v[7] << 7) | (v[8] << 2) | (v[9] >> 3)) & 255;
    result_bytes[6] := ((v[9] << 5) | v[10]) & 255;

    -- Decode entropy (80 bits) - following Go implementation exactly with proper bit masking
    result_bytes[7] := ((v[11] << 3) | (v[12] >> 2)) & 255;
    result_bytes[8] := ((v[12] << 6) | (v[13] << 1) | (v[14] >> 4)) & 255;
    result_bytes[9] := ((v[14] << 4) | (v[15] >> 1)) & 255;
    result_bytes[10] := ((v[15] << 7) | (v[16] << 2) | (v[17] >> 3)) & 255;
    result_bytes[11] := ((v[17] << 5) | v[18]) & 255;
    result_bytes[12] := ((v[19] << 3) | (v[20] >> 2)) & 255;
    result_bytes[13] := ((v[20] << 6) | (v[21] << 1) | (v[22] >> 4)) & 255;
    result_bytes[14] := ((v[22] << 4) | (v[23] >> 1)) & 255;
    result_bytes[15] := ((v[23] << 7) | (v[24] << 2) | (v[25] >> 3)) & 255;
    result_bytes[16] := ((v[25] << 5) | v[26]) & 255;

    -- Convert result bytes to hex string directly
    hex_string := '';
    FOR i IN 1..16
        LOOP
            -- Ensure byte values are in valid range (0-255)
            IF result_bytes[i] < 0 OR result_bytes[i] > 255 THEN
                RAISE EXCEPTION 'Invalid byte value % at position % during ULID decoding', result_bytes[i], i
                    USING ERRCODE = '22023'; -- invalid_parameter_value
            END IF;

            -- Convert byte to 2-character hex string with zero padding
            hex_string := hex_string || lpad(to_hex(result_bytes[i]), 2, '0');
        END LOOP;

    -- Format as UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
    RETURN (substr(hex_string, 1, 8) || '-' ||
            substr(hex_string, 9, 4) || '-' ||
            substr(hex_string, 13, 4) || '-' ||
            substr(hex_string, 17, 4) || '-' ||
            substr(hex_string, 21, 12))::UUID;

END;
$$;

-- Create a comment explaining the ulid_to_uuid function
COMMENT ON FUNCTION ulid_to_uuid(TEXT) IS
    'Converts a ULID string back to its corresponding UUID representation. '
    'Based on pixie-sh/ulid-go implementation with Crockford Base32 decoding. '
    'Raises exceptions for invalid ULID format or conversion errors.';

-- Function to explode a pULID (UUID format) into its components
CREATE OR REPLACE FUNCTION pulid_explode_uuid(input_uuid UUID)
    RETURNS TABLE
            (
                ulid            TEXT,
                uuid            UUID,
                scope           INTEGER,
                timestamp_epoch BIGINT
            )
    LANGUAGE plpgsql
    IMMUTABLE
    STRICT
    SECURITY DEFINER
AS
$$
DECLARE
    binary_data         BYTEA;
    uuid_text           TEXT;
    bytes               INT[16];
    extracted_scope     INTEGER;
    extracted_timestamp BIGINT;
    ulid_string         TEXT;
BEGIN
    -- Validate input
    IF input_uuid IS NULL THEN
        RETURN;
    END IF;

    -- Convert UUID to text and validate format
    uuid_text := input_uuid::TEXT;

    -- Validate UUID format (8-4-4-4-12 pattern with hyphens)
    IF NOT uuid_text ~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$' THEN
        RAISE EXCEPTION 'Invalid UUID format: %', uuid_text
            USING ERRCODE = '22023',
                HINT = 'UUID must be in format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
    END IF;

    -- Convert UUID to binary
    BEGIN
        binary_data := decode(replace(uuid_text, '-', ''), 'hex');
    EXCEPTION
        WHEN OTHERS THEN
            RAISE EXCEPTION 'Failed to decode UUID hex data: %', uuid_text
                USING ERRCODE = '22023',
                    HINT = 'UUID contains invalid hexadecimal characters';
    END;

    -- Validate binary data length
    IF length(binary_data) != 16 THEN
        RAISE EXCEPTION 'Invalid UUID binary length: expected 16 bytes, got %', length(binary_data)
            USING ERRCODE = '22023';
    END IF;

    -- Extract individual bytes
    FOR i IN 1..16
        LOOP
            bytes[i] := get_byte(binary_data, i - 1);
        END LOOP;

    -- Extract timestamp (first 6 bytes) - convert to milliseconds since epoch
    -- Following the Go implementation: bytes 0-5 represent timestamp
    extracted_timestamp := (
        (bytes[1]::BIGINT << 40) |
        (bytes[2]::BIGINT << 32) |
        (bytes[3]::BIGINT << 24) |
        (bytes[4]::BIGINT << 16) |
        (bytes[5]::BIGINT << 8) |
        bytes[6]::BIGINT
        );

    -- Extract scope (bytes 6-7) - 2 bytes for scope
    -- Following the Go implementation: bytes 6-7 represent scope (uint16, big-endian)
    extracted_scope := (
        (bytes[7] << 8) | bytes[8]
        );

    -- Convert to ULID string format
    ulid_string := uuid_to_ulid(input_uuid);

    -- Return the exploded components
    RETURN QUERY SELECT ulid_string,
                        input_uuid,
                        extracted_scope,
                        extracted_timestamp;
END;
$$;

-- Create a comment explaining the pulid_explode function
COMMENT ON FUNCTION pulid_explode_uuid(UUID) IS
    'Explodes a pULID (UUID format) into its constituent components: ULID string, UUID, scope, and timestamp epoch (milliseconds). '
    'Based on pixie-sh/ulid-go pULID structure: 6 bytes timestamp + 2 bytes scope + 8 bytes entropy. '
    'Returns a table with columns: ulid (TEXT), uuid (UUID), scope (INTEGER), timestamp_epoch (BIGINT).';

-- Convenience function to explode from ULID string format
CREATE OR REPLACE FUNCTION pulid_explode_ulid(input_ulid TEXT)
    RETURNS TABLE
            (
                ulid            TEXT,
                uuid            UUID,
                scope           INTEGER,
                timestamp_epoch BIGINT
            )
    LANGUAGE plpgsql
    IMMUTABLE
    STRICT
    SECURITY DEFINER
AS
$$
DECLARE
    converted_uuid UUID;
BEGIN
    -- Convert ULID to UUID first
    converted_uuid := ulid_to_uuid(input_ulid);

    -- Use the main explode function
    RETURN QUERY SELECT * FROM pulid_explode_uuid(converted_uuid);
END;
$$;

-- Create a comment explaining the pulid_explode_from_ulid function
COMMENT ON FUNCTION pulid_explode_ulid(TEXT) IS
    'Explodes a ULID string into its constituent components by first converting to UUID format, then extracting components. '
    'Returns a table with columns: ulid (TEXT), uuid (UUID), scope (INTEGER), timestamp_epoch (BIGINT).';

-- Comprehensive test cases including error scenarios
DO
$$
    DECLARE
        test_uuid      UUID;
        test_ulid      TEXT;
        test_result    TEXT;
        converted_uuid UUID;
        error_msg      TEXT;
    BEGIN
        RAISE NOTICE '=== STARTING COMPREHENSIVE TESTS ===';

        -- Test 1: Basic UUID to ULID conversion
        test_uuid := '0194aa0e-dcde-ffff-53f2-18257f60b037'::UUID;
        test_result := uuid_to_ulid(test_uuid);
        RAISE NOTICE 'Test 1 - UUID to ULID: % -> %', test_uuid, test_result;

        -- Test 2: Another UUID to ULID conversion
        test_uuid := '0194aa15-34ab-0237-22e7-0b92b8c5f1b7'::UUID;
        test_result := uuid_to_ulid(test_uuid);
        RAISE NOTICE 'Test 2 - UUID to ULID: % -> %', test_uuid, test_result;

        -- Test 3: Round-trip conversion (UUID -> ULID -> UUID)
        test_uuid := '0194aa0e-dcde-ffff-53f2-18257f60b037'::UUID;
        test_ulid := uuid_to_ulid(test_uuid);
        converted_uuid := ulid_to_uuid(test_ulid);
        RAISE NOTICE 'Test 3 - Round-trip: % -> % -> %', test_uuid, test_ulid, converted_uuid;
        IF test_uuid = converted_uuid THEN
            RAISE NOTICE 'Test 3 - Round-trip: SUCCESS ✓';
        ELSE
            RAISE NOTICE 'Test 3 - Round-trip: FAILED ✗ (% != %)', test_uuid, converted_uuid;
        END IF;

        -- Test 4: ULID to UUID conversion
        test_ulid := '01HQZD7VZFN777MGPKM0YG1YJX';
        converted_uuid := ulid_to_uuid(test_ulid);
        RAISE NOTICE 'Test 4 - ULID to UUID: % -> %', test_ulid, converted_uuid;

        -- Test 5: NULL handling
        test_result := uuid_to_ulid(NULL);
        RAISE NOTICE 'Test 5 - NULL UUID -> ULID: %', COALESCE(test_result, 'NULL');

        converted_uuid := ulid_to_uuid(NULL);
        RAISE NOTICE 'Test 5 - NULL ULID -> UUID: %', COALESCE(converted_uuid::TEXT, 'NULL');

        RAISE NOTICE '=== ERROR HANDLING TESTS ===';

        -- Test 6: Invalid ULID length (too short)
        BEGIN
            converted_uuid := ulid_to_uuid('TOO_SHORT');
            RAISE NOTICE 'Test 6 - FAILED: Should have thrown error for short ULID';
        EXCEPTION
            WHEN OTHERS THEN
                GET STACKED DIAGNOSTICS error_msg = MESSAGE_TEXT;
                RAISE NOTICE 'Test 6 - SHORT ULID ERROR: % ✓', error_msg;
        END;

        -- Test 7: Invalid ULID length (too long)
        BEGIN
            converted_uuid := ulid_to_uuid('01HQZD7VZFN777MGPKM0YG1YJXTOOLONG');
            RAISE NOTICE 'Test 7 - FAILED: Should have thrown error for long ULID';
        EXCEPTION
            WHEN OTHERS THEN
                GET STACKED DIAGNOSTICS error_msg = MESSAGE_TEXT;
                RAISE NOTICE 'Test 7 - LONG ULID ERROR: % ✓', error_msg;
        END;

        -- Test 8: Invalid ULID characters
        BEGIN
            converted_uuid := ulid_to_uuid('01HQZD7VZFN777MGPKM0YG1YJI'); -- 'I' is invalid in Crockford Base32
            RAISE NOTICE 'Test 8 - FAILED: Should have thrown error for invalid char';
        EXCEPTION
            WHEN OTHERS THEN
                GET STACKED DIAGNOSTICS error_msg = MESSAGE_TEXT;
                RAISE NOTICE 'Test 8 - INVALID CHAR ERROR: % ✓', error_msg;
        END;

        -- Test 9: ULID timestamp overflow (first char > 7)
        BEGIN
            converted_uuid := ulid_to_uuid('81HQZD7VZFN777MGPKM0YG1YJX'); -- Starts with '8' which is too large
            RAISE NOTICE 'Test 9 - FAILED: Should have thrown error for timestamp overflow';
        EXCEPTION
            WHEN OTHERS THEN
                GET STACKED DIAGNOSTICS error_msg = MESSAGE_TEXT;
                RAISE NOTICE 'Test 9 - TIMESTAMP OVERFLOW ERROR: % ✓', error_msg;
        END;

        -- Test 10: Multiple round-trip conversions to verify consistency
        RAISE NOTICE '=== CONSISTENCY TESTS ===';
        FOR i IN 1..5
            LOOP
                -- Generate different test UUIDs with proper format (8-4-4-4-12)
                test_uuid := (substr(md5(i::TEXT), 1, 8) || '-' ||
                              substr(md5(i::TEXT), 9, 4) || '-' ||
                              substr(md5(i::TEXT), 13, 4) || '-' ||
                              substr(md5(i::TEXT), 17, 4) || '-' ||
                              substr(md5((i * 2)::TEXT), 1, 12))::UUID;
                test_ulid := uuid_to_ulid(test_uuid);
                converted_uuid := ulid_to_uuid(test_ulid);

                IF test_uuid = converted_uuid THEN
                    RAISE NOTICE 'Test 10.% - Consistency: % ✓', i, test_uuid;
                ELSE
                    RAISE NOTICE 'Test 10.% - Consistency: FAILED ✗ % != %', i, test_uuid, converted_uuid;
                END IF;
            END LOOP;

        RAISE NOTICE '=== ALL TESTS COMPLETED ===';

    END
$$;

-- Test cases for pulid_explode functions
DO
$$
    DECLARE
        test_uuid      UUID;
        test_ulid      TEXT;
        explode_result RECORD;
    BEGIN
        RAISE NOTICE '=== STARTING PULID_EXPLODE TESTS ===';

        -- Test 1: Explode pULID from UUID format
        test_uuid := '0194aa0e-dcde-ffff-53f2-18257f60b037'::UUID;
        SELECT * INTO explode_result FROM pulid_explode_uuid(test_uuid);
        RAISE NOTICE 'Test 1 - pULID Explode from UUID:';
        RAISE NOTICE '  Input UUID: %', test_uuid;
        RAISE NOTICE '  ULID: %', explode_result.ulid;
        RAISE NOTICE '  UUID: %', explode_result.uuid;
        RAISE NOTICE '  Scope: %', explode_result.scope;
        RAISE NOTICE '  Timestamp (ms): %', explode_result.timestamp_epoch;
        RAISE NOTICE '  Timestamp (ISO): %', to_timestamp(explode_result.timestamp_epoch / 1000.0);

        -- Test 2: Explode pULID from ULID string format
        test_ulid := '01HQZD7VZFN777MGPKM0YG1YJX';
        SELECT * INTO explode_result FROM pulid_explode_ulid(test_ulid);
        RAISE NOTICE 'Test 2 - pULID Explode from ULID string:';
        RAISE NOTICE '  Input ULID: %', test_ulid;
        RAISE NOTICE '  ULID: %', explode_result.ulid;
        RAISE NOTICE '  UUID: %', explode_result.uuid;
        RAISE NOTICE '  Scope: %', explode_result.scope;
        RAISE NOTICE '  Timestamp (ms): %', explode_result.timestamp_epoch;
        RAISE NOTICE '  Timestamp (ISO): %', to_timestamp(explode_result.timestamp_epoch / 1000.0);

        -- Test 3: Explode another example
        test_uuid := '0194aa15-34ab-0237-22e7-0b92b8c5f1b7'::UUID;
        SELECT * INTO explode_result FROM pulid_explode_uuid(test_uuid);
        RAISE NOTICE 'Test 3 - Another pULID Explode:';
        RAISE NOTICE '  Input UUID: %', test_uuid;
        RAISE NOTICE '  ULID: %', explode_result.ulid;
        RAISE NOTICE '  UUID: %', explode_result.uuid;
        RAISE NOTICE '  Scope: %', explode_result.scope;
        RAISE NOTICE '  Timestamp (ms): %', explode_result.timestamp_epoch;
        RAISE NOTICE '  Timestamp (ISO): %', to_timestamp(explode_result.timestamp_epoch / 1000.0);

        RAISE NOTICE '=== PULID_EXPLODE TESTS COMPLETED ===';

    END
$$;

-- Optional: Create an operator for convenience
-- CREATE OPERATOR ::ulid (
--     FUNCTION = uuid_to_ulid,
--     LEFTARG = UUID
-- );
