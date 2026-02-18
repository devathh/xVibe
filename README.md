# xVibe - cloud messenger
## 👋 About
xVibe is a secure enterprise-level messenger that ensures the confidentiality of your correspondence through a multi-level encryption system. All data is transmitted over a secure TLS 1.3 channel, ensuring the integrity and uninterceptability of your traffic at the network level.
The Envelope Encryption architecture is used to store messages, utilizing the AES-GCM algorithm. Each chat is protected by a unique data encryption key (DEK), which is stored in the database only in encrypted form using a master key (KEK). This hierarchical model ensures that even if the storage is compromised, the data remains inaccessible without access to the managed server keys.

## 💻 Tech Stack
- jwt authorization
- postgres - permanent storage
- redis - caching
- nginx - routing

## 🏁 Quick Start
1. To get started quickly, you need to clone the project: `git clone https://github.com/devathh/xVibe.git`
2. Then, to directly launch the application, it is better to use docker. To launch the compost, use `make quickstart`
3. enjoy :)
