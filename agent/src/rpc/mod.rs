/*
 * Copyright (c) 2024 KhulnaSoft, Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

mod ntp;
mod session;
mod synchronizer;

pub use session::{Session, DEFAULT_TIMEOUT};
pub(crate) use synchronizer::{StaticConfig, Status, Synchronizer};

cfg_if::cfg_if! {
    if #[cfg(any(target_os = "linux", target_os = "android"))] {
        pub mod remote_exec;
        pub use remote_exec::Executor;
    }
}

use std::time::{Duration, SystemTime};

const RPC_RETRY_INTERVAL: Duration = Duration::from_secs(60);

pub fn get_timestamp(ntp_diff: i64) -> Duration {
    let now = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_nanos() as i128
        + ntp_diff as i128;
    Duration::from_nanos(now as u64)
}
