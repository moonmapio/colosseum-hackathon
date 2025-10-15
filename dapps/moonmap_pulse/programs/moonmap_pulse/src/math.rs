// Prevents overflow by stopping the sum once the cap is reached.
pub fn safe_sum_with_u16cap(values: &[u16], max: u16) -> u16 {
    let mut total: u16 = 0;
    for &v in values {
        total = total.saturating_add(v);
        if total >= max {
            return max;
        }
    }
    total
}
