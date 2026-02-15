#!/bin/bash

# ShizuMusic Session Validator & Generator
# Fixes invalid/corrupted sessions

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔐 ShizuMusic Session Validator & Fixer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if STRING_SESSION exists in .env
if [ -f .env ]; then
    SESSION=$(grep "STRING_SESSION=" .env | cut -d'=' -f2)
    if [ -z "$SESSION" ]; then
        echo "⚠️  No STRING_SESSION found in .env"
    else
        echo "✅ Found STRING_SESSION in .env"
        echo "   Length: ${#SESSION} characters"
        echo ""
        
        # Validate session format
        if [[ $SESSION =~ ^1[A-Za-z0-9_-]+$ ]]; then
            echo "✅ Session format looks valid"
        else
            echo "❌ Session format invalid!"
            echo "   Session should start with '1' followed by base64 chars"
        fi
    fi
else
    echo "❌ .env file not found!"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "💡 Options to Fix:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1️⃣  Generate NEW session (Recommended)"
echo "2️⃣  Use Python Telethon to generate"
echo "3️⃣  Use Python Pyrogram to generate"
echo "4️⃣  Skip user session (bot only mode)"
echo ""
read -p "Enter choice (1-4): " choice

case $choice in
    1)
        echo ""
        echo "🔧 Installing Python dependencies..."
        pip3 install telethon --break-system-packages --quiet 2>/dev/null || pip3 install telethon --quiet
        
        echo "✅ Ready to generate session!"
        echo ""
        
        python3 << 'EOFPYTHON'
from telethon.sync import TelegramClient
from telethon.sessions import StringSession
import sys

print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print("📱 Session Generator (Telethon)")
print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print()

try:
    api_id = int(input("Enter API_ID: "))
    api_hash = input("Enter API_HASH: ")
    
    print()
    print("⏳ Connecting to Telegram...")
    print("   Please enter your phone number and verification code")
    print()
    
    with TelegramClient(StringSession(), api_id, api_hash) as client:
        session_string = client.session.save()
        
        print()
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        print("✅ NEW Session Generated Successfully!")
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        print()
        print("📝 Your NEW STRING_SESSION:")
        print()
        print(session_string)
        print()
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        print()
        
        # Save to file
        save = input("💾 Update .env file automatically? (y/n): ").strip().lower()
        if save in ['y', 'yes']:
            # Update .env
            import os
            import re
            
            if os.path.exists('.env'):
                with open('.env', 'r') as f:
                    content = f.read()
                
                # Replace STRING_SESSION
                if 'STRING_SESSION=' in content:
                    content = re.sub(r'STRING_SESSION=.*', f'STRING_SESSION={session_string}', content)
                else:
                    content += f'\nSTRING_SESSION={session_string}\n'
                
                with open('.env', 'w') as f:
                    f.write(content)
                
                print("✅ .env updated!")
            else:
                with open('.env', 'a') as f:
                    f.write(f'STRING_SESSION={session_string}\n')
                print("✅ Added to .env")
        else:
            print()
            print("📝 Copy this to your .env file manually:")
            print(f"STRING_SESSION={session_string}")
        
        print()
        print("✅ Done! Now run: ./shizumusic")
        
except KeyboardInterrupt:
    print("\n\n❌ Cancelled by user")
    sys.exit(1)
except Exception as e:
    print(f"\n\n❌ Error: {e}")
    sys.exit(1)
EOFPYTHON
        ;;
        
    2)
        echo ""
        echo "📦 Installing Telethon..."
        pip3 install telethon --break-system-packages --quiet 2>/dev/null || pip3 install telethon --quiet
        
        python3 -c "from telethon.sync import TelegramClient; from telethon.sessions import StringSession; api_id=int(input('API_ID: ')); api_hash=input('API_HASH: '); print('\n⏳ Connecting...\n'); client=TelegramClient(StringSession(),api_id,api_hash); client.start(); print('\n✅ Session:\n'); print(client.session.save()); print()"
        ;;
        
    3)
        echo ""
        echo "📦 Installing Pyrogram..."
        pip3 install pyrogram tgcrypto --break-system-packages --quiet 2>/dev/null || pip3 install pyrogram tgcrypto --quiet
        
        python3 -c "from pyrogram import Client; import os; api_id=int(input('API_ID: ')); api_hash=input('API_HASH: '); print('\n⏳ Connecting...\n'); app=Client('temp',api_id,api_hash); app.start(); print('\n✅ Session:\n'); print(app.export_session_string()); app.stop(); os.remove('temp.session')"
        ;;
        
    4)
        echo ""
        echo "⚠️  Running in bot-only mode (no voice chat support)"
        echo ""
        echo "To disable user client temporarily:"
        echo "  1. Edit .env"
        echo "  2. Comment out or remove STRING_SESSION"
        echo "  3. Restart bot"
        echo ""
        echo "Note: Voice chat streaming will not work without user session!"
        ;;
        
    *)
        echo "❌ Invalid choice!"
        exit 1
        ;;
esac

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
