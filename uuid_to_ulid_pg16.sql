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
AS $$
DECLARE
    -- Base32 encoding table used by ULID (Crockford's base32)
    encoding CONSTANT TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    
    -- Binary representation of the UUID
    binary_data BYTEA;
    
    -- Individual bytes for processing
    b0 INT; b1 INT; b2 INT; b3 INT; b4 INT; b5 INT;
    b6 INT; b7 INT; b8 INT; b9 INT; b10 INT; b11 INT;
    b12 INT; b13 INT; b14 INT; b15 INT;
    
    -- Result ULID string
    result TEXT := '';
    
BEGIN
    -- Validate input
    IF input_uuid IS NULL THEN
        RETURN NULL;
    END IF;
    
    -- Convert UUID to binary (remove hyphens and decode hex)
    binary_data := decode(replace(input_uuid::TEXT, '-', ''), 'hex');
    
    -- Extract individual bytes
    b0 := get_byte(binary_data, 0);
    b1 := get_byte(binary_data, 1);
    b2 := get_byte(binary_data, 2);
    b3 := get_byte(binary_data, 3);
    b4 := get_byte(binary_data, 4);
    b5 := get_byte(binary_data, 5);
    b6 := get_byte(binary_data, 6);
    b7 := get_byte(binary_data, 7);
    b8 := get_byte(binary_data, 8);
    b9 := get_byte(binary_data, 9);
    b10 := get_byte(binary_data, 10);
    b11 := get_byte(binary_data, 11);
    b12 := get_byte(binary_data, 12);
    b13 := get_byte(binary_data, 13);
    b14 := get_byte(binary_data, 14);
    b15 := get_byte(binary_data, 15);
    
    -- Encode timestamp part (first 10 characters of ULID - bytes 0-5)
    -- Following the exact bit manipulation from the Go implementation
    result := result || substr(encoding, ((b0 & 224) >> 5) + 1, 1);        -- dst[0]
    result := result || substr(encoding, (b0 & 31) + 1, 1);                -- dst[1]
    result := result || substr(encoding, ((b1 & 248) >> 3) + 1, 1);        -- dst[2]
    result := result || substr(encoding, (((b1 & 7) << 2) | ((b2 & 192) >> 6)) + 1, 1); -- dst[3]
    result := result || substr(encoding, ((b2 & 62) >> 1) + 1, 1);         -- dst[4]
    result := result || substr(encoding, (((b2 & 1) << 4) | ((b3 & 240) >> 4)) + 1, 1); -- dst[5]
    result := result || substr(encoding, (((b3 & 15) << 1) | ((b4 & 128) >> 7)) + 1, 1); -- dst[6]
    result := result || substr(encoding, ((b4 & 124) >> 2) + 1, 1);        -- dst[7]
    result := result || substr(encoding, (((b4 & 3) << 3) | ((b5 & 224) >> 5)) + 1, 1); -- dst[8]
    result := result || substr(encoding, (b5 & 31) + 1, 1);                -- dst[9]
    
    -- Encode entropy part (last 16 characters of ULID - bytes 6-15)
    result := result || substr(encoding, ((b6 & 248) >> 3) + 1, 1);        -- dst[10]
    result := result || substr(encoding, (((b6 & 7) << 2) | ((b7 & 192) >> 6)) + 1, 1); -- dst[11]
    result := result || substr(encoding, ((b7 & 62) >> 1) + 1, 1);         -- dst[12]
    result := result || substr(encoding, (((b7 & 1) << 4) | ((b8 & 240) >> 4)) + 1, 1); -- dst[13]
    result := result || substr(encoding, (((b8 & 15) << 1) | ((b9 & 128) >> 7)) + 1, 1); -- dst[14]
    result := result || substr(encoding, ((b9 & 124) >> 2) + 1, 1);        -- dst[15]
    result := result || substr(encoding, (((b9 & 3) << 3) | ((b10 & 224) >> 5)) + 1, 1); -- dst[16]
    result := result || substr(encoding, (b10 & 31) + 1, 1);               -- dst[17]
    result := result || substr(encoding, ((b11 & 248) >> 3) + 1, 1);       -- dst[18]
    result := result || substr(encoding, (((b11 & 7) << 2) | ((b12 & 192) >> 6)) + 1, 1); -- dst[19]
    result := result || substr(encoding, ((b12 & 62) >> 1) + 1, 1);        -- dst[20]
    result := result || substr(encoding, (((b12 & 1) << 4) | ((b13 & 240) >> 4)) + 1, 1); -- dst[21]
    result := result || substr(encoding, (((b13 & 15) << 1) | ((b14 & 128) >> 7)) + 1, 1); -- dst[22]
    result := result || substr(encoding, ((b14 & 124) >> 2) + 1, 1);       -- dst[23]
    result := result || substr(encoding, (((b14 & 3) << 3) | ((b15 & 224) >> 5)) + 1, 1); -- dst[24]
    result := result || substr(encoding, (b15 & 31) + 1, 1);               -- dst[25]
    
    RETURN result;
    
EXCEPTION
    WHEN OTHERS THEN
        -- Return NULL for any conversion errors
        RETURN NULL;
END;
$$;

-- Create a comment explaining the function
COMMENT ON FUNCTION uuid_to_ulid(UUID) IS 
    'Converts a UUID that follows pULID format to ULID string representation. ' ||
    'Based on pixie-sh/ulid-go implementation with Crockford Base32 encoding. ' ||
    'Returns NULL for invalid UUIDs or conversion errors.';

-- Example usage and test cases
DO $$
DECLARE
    test_uuid UUID;
    test_result TEXT;
BEGIN
    -- Test with example from README
    test_uuid := '0194aa0e-dcde-ffff-53f2-18257f60b037'::UUID;
    test_result := uuid_to_ulid(test_uuid);
    RAISE NOTICE 'UUID: % -> ULID: %', test_uuid, test_result;
    
    -- Test with another example 
    test_uuid := '0194aa15-34ab-0237-22e7-0b92b8c5f1b7'::UUID;
    test_result := uuid_to_ulid(test_uuid);
    RAISE NOTICE 'UUID: % -> ULID: %', test_uuid, test_result;
    
    -- Test with NULL
    test_result := uuid_to_ulid(NULL);
    RAISE NOTICE 'NULL UUID -> ULID: %', COALESCE(test_result, 'NULL');
    
END $$;

-- Optional: Create an operator for convenience
-- CREATE OPERATOR ::ulid (
--     FUNCTION = uuid_to_ulid,
--     LEFTARG = UUID
-- );
