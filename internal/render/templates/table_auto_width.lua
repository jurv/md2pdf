-- md2pdf table auto-width filter
--
-- Pandoc leaves many Markdown tables with default column widths, which the
-- LaTeX writer then turns into equal-width p-columns. That is often a poor fit
-- for real-world documentation tables where one or two descriptive columns need
-- much more space than short status/boolean columns.
--
-- This filter assigns relative column widths when the table has no explicit
-- widths yet. Two complementary mechanisms drive the result:
--   * a per-column floor derived from the longest unbreakable token in that
--     column, so a column never receives less than what its widest word
--     physically requires (LaTeX cannot hyphenate mid-token without language
--     support, so a too-narrow column produces visible overflow);
--   * a score-based distribution of the remaining width, balancing the length
--     of prose content against the length of unbreakable tokens.

local text = pandoc.text

local function text_len(value)
  if text and text.len then
    return text.len(value)
  end
  return #value
end

local function normalize_space(value)
  value = value:gsub("%s+", " ")
  value = value:gsub("^%s+", "")
  value = value:gsub("%s+$", "")
  return value
end

local function longest_token_len(value)
  local longest = 0
  for token in value:gmatch("%S+") do
    token = token:gsub("^[%p]+", "")
    token = token:gsub("[%p]+$", "")
    local length = text_len(token)
    if length > longest then
      longest = length
    end
  end
  return longest
end

local function cell_metrics(contents)
  local value = normalize_space(pandoc.utils.stringify(contents))
  if value == "" then
    return 0, 0
  end
  return text_len(value), longest_token_len(value)
end

local function cell_score(compact_len, token_len)
  -- Compact length captures the demand of long prose; token length captures
  -- the demand of unbreakable words (which drive overflow when a column is
  -- narrower than the widest token). The two coefficients keep their relative
  -- balance similar to the historical heuristic, but at higher absolute weight
  -- so the resulting distribution better matches physical width requirements.
  local score = math.max(compact_len * 0.85, token_len * 3.5)
  return math.max(6, score)
end

local function token_floor(longest)
  if longest <= 0 then
    return 0.05
  end
  -- ~1.1% per character + a 4% base, capped so a single very long token cannot
  -- single-handedly claim a third of the page width when other columns also
  -- need room. Floors are normalized further below if their sum is too high.
  return math.max(0.05, math.min(0.30, 0.011 * longest + 0.04))
end

local function accumulate_row(metrics, row)
  for column_index, cell in ipairs(row.cells) do
    if metrics[column_index] then
      local compact, token = cell_metrics(cell.contents)
      if compact > metrics[column_index].compact then
        metrics[column_index].compact = compact
      end
      if token > metrics[column_index].token then
        metrics[column_index].token = token
      end
    end
  end
end

local function should_auto_size(tbl)
  -- Always recompute widths for tables with at least two columns. Pandoc's
  -- pipe-table syntax derives widths from the number of dashes in the
  -- separator line, which authors typically use to visually align their
  -- markdown rather than to express deliberate column proportions; honoring
  -- those widths produces unbalanced output. Authors who do want fixed widths
  -- should rely on grid-table syntax or post-process via their own filter.
  return #tbl.colspecs >= 2
end

function Table(tbl)
  if not should_auto_size(tbl) then
    return tbl
  end

  local column_count = #tbl.colspecs
  local metrics = {}
  for i = 1, column_count do
    metrics[i] = {compact = 0, token = 0}
  end

  for _, row in ipairs(tbl.head.rows) do
    accumulate_row(metrics, row)
  end

  for _, body in ipairs(tbl.bodies) do
    for _, row in ipairs(body.head) do
      accumulate_row(metrics, row)
    end
    for _, row in ipairs(body.body) do
      accumulate_row(metrics, row)
    end
  end

  for _, row in ipairs(tbl.foot.rows) do
    accumulate_row(metrics, row)
  end

  -- Per-column floor, derived from the longest unbreakable token observed.
  local floors = {}
  local floor_sum = 0
  for i = 1, column_count do
    floors[i] = token_floor(metrics[i].token)
    floor_sum = floor_sum + floors[i]
  end
  -- Cap the total floor budget so the score-based distribution still has
  -- meaningful room to differentiate columns with similar floors.
  local floor_budget = 0.55
  if floor_sum > floor_budget then
    local k = floor_budget / floor_sum
    for i = 1, column_count do
      floors[i] = floors[i] * k
    end
    floor_sum = floor_budget
  end

  local scores = {}
  local total_score = 0
  for i = 1, column_count do
    scores[i] = cell_score(metrics[i].compact, metrics[i].token)
    total_score = total_score + scores[i]
  end
  if total_score <= 0 then
    return tbl
  end

  local distributable = 1 - floor_sum
  if distributable < 0 then
    distributable = 0
  end

  local colspecs = {}
  for i, colspec in ipairs(tbl.colspecs) do
    local width = floors[i] + distributable * (scores[i] / total_score)
    colspecs[i] = {colspec[1], width}
  end
  tbl.colspecs = colspecs

  return tbl
end
