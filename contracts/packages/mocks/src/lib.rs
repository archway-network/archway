#[cfg(not(target_arch = "wasm32"))]
mod querier;

#[cfg(not(target_arch = "wasm32"))]
pub use querier::{mock_dependencies, MockQuerier};
