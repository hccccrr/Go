#!/usr/bin/env python3
"""
Session Converter for Telegram
Converts Pyrogram/Telethon sessions to Gogram format
"""

import base64
import struct
import sys

def decode_pyrogram_session(session_string):
    """Decode Pyrogram session string"""
    # Add padding if needed
    while len(session_string) % 4 != 0:
        session_string += "="
    
    try:
        # Pyrogram uses URL-safe base64
        packed_data = base64.urlsafe_b64decode(session_string)
    except Exception as e:
        raise ValueError(f"Failed to decode base64: {e}")
    
    # Expected length: 271 bytes
    if len(packed_data) != 271:
        raise ValueError(f"Invalid Pyrogram session length: {len(packed_data)} (expected 271)")
    
    # Extract fields
    dc_id = struct.unpack('B', packed_data[0:1])[0]
    test_mode = packed_data[5] != 0
    auth_key = packed_data[6:262]
    
    # Map DC to hostname
    dc_map_prod = {
        1: "149.154.175.53:443",
        2: "149.154.167.51:443",
        3: "149.154.175.100:443",
        4: "149.154.167.91:443",
        5: "91.108.56.130:443"
    }
    
    dc_map_test = {
        1: "149.154.175.10:443",
        2: "149.154.167.40:443",
        3: "149.154.175.117:443"
    }
    
    dc_map = dc_map_test if test_mode else dc_map_prod
    hostname = dc_map.get(dc_id, f"Unknown DC {dc_id}")
    
    return hostname, auth_key

def decode_telethon_session(session_string):
    """Decode Telethon session string"""
    if not session_string.startswith('1'):
        raise ValueError("Invalid Telethon session: must start with '1'")
    
    # Remove '1' prefix and decode
    data = base64.urlsafe_b64decode(session_string[1:])
    
    # Determine IP length
    ip_len = 4 if len(data) == 263 else 16
    expected_len = 1 + ip_len + 2 + 256
    
    if len(data) != expected_len:
        raise ValueError(f"Invalid Telethon session length: {len(data)} (expected {expected_len})")
    
    offset = 1
    
    # Extract IP
    ip_data = data[offset:offset+ip_len]
    if ip_len == 4:
        ip = '.'.join(str(b) for b in ip_data)
    else:
        ip = ':'.join(f'{ip_data[i]:02x}{ip_data[i+1]:02x}' for i in range(0, 16, 2))
    offset += ip_len
    
    # Extract port (big-endian)
    port = struct.unpack('>H', data[offset:offset+2])[0]
    offset += 2
    
    # Extract auth key
    auth_key = data[offset:offset+256]
    
    hostname = f"{ip}:{port}"
    return hostname, auth_key

def create_gogram_session(hostname, auth_key):
    """Create Gogram session string"""
    # Gogram session format (as per gogram source):
    # hostname_len (varint) + hostname + auth_key
    
    hostname_bytes = hostname.encode('utf-8')
    hostname_len = len(hostname_bytes)
    
    # Create session data
    session_data = bytearray()
    
    # Write varint for hostname length
    while hostname_len >= 0x80:
        session_data.append((hostname_len & 0x7F) | 0x80)
        hostname_len >>= 7
    session_data.append(hostname_len & 0x7F)
    
    # Add hostname
    session_data.extend(hostname_bytes)
    
    # Add auth key
    session_data.extend(auth_key)
    
    # Encode as base64
    return base64.urlsafe_b64encode(bytes(session_data)).decode('utf-8').rstrip('=')

def convert_session(session_string):
    """Auto-detect and convert session format"""
    session_string = session_string.strip()
    
    # Detect format
    if session_string.startswith('1'):
        print("🔍 Detected: Telethon session")
        hostname, auth_key = decode_telethon_session(session_string)
    else:
        print("🔍 Detected: Pyrogram session")
        hostname, auth_key = decode_pyrogram_session(session_string)
    
    print(f"📡 Hostname: {hostname}")
    print(f"🔑 Auth Key: {len(auth_key)} bytes")
    print()
    
    # Convert to Gogram format
    gogram_session = create_gogram_session(hostname, auth_key)
    print("✅ Converted to Gogram format!")
    print()
    print("=" * 80)
    print("GOGRAM SESSION STRING:")
    print("=" * 80)
    print(gogram_session)
    print("=" * 80)
    
    return gogram_session

if __name__ == "__main__":
    if len(sys.argv) > 1:
        session = sys.argv[1]
    else:
        # Use the session from your env file
        session = "AgGIzloAS4zC-M9OIJYQfFRHYO0mGR81rqYFx3v9AKqhi3qRZvinIIP3xeif7YiitdzoVwtDX5P8U_XPkl91ZDmcX8MvhxSgFZ02Z5VKOuWF4eEZOJr9zFOR9ZH7xEdbbah58cS3OsyaVyuiJdeb94n5WmpHQSM0jR4Ciiprj4OCdlHFyRfxnUdU6_A1M8_C-QXFcFnrybuCAtV1ITPk4WQdA2qbCghSRww47m33skNzne50KxTzB811-Nbs2lt_rIl3sqqmfRzfDg4ukLgSKbFLw1uR3EmjsgPU-fyzV-_7d3EdFEZJ26pUA39rV8vPuc3pEhIs1L0zmCXNwfNaGurCgoMCMwAAAAHjG_iTAA"
    
    print("🔄 Telegram Session Converter")
    print("=" * 80)
    print()
    
    try:
        gogram_session = convert_session(session)
        
        print()
        print("📝 Copy the session string above and use it in your .env file")
        print("   Replace the current STRING_SESSION value with the new one")
        
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)
