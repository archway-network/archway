#[cfg(not(target_arch = "wasm32"))]
pub mod querier;

#[cfg(not(target_arch = "wasm32"))]
pub use querier::{MockQuerier, mock_dependencies_with_wasm_query};
