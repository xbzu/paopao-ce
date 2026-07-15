<template>
    <div>
        <main-nav title="管理后台" />

        <n-alert
            v-if="loadError"
            title="管理数据加载失败"
            type="error"
            class="load-error-alert"
        >
            <div class="load-error-content">
                <span>{{ loadError }}</span>
                <n-button size="small" type="error" secondary @click="loadSettings">
                    重新加载
                </n-button>
            </div>
        </n-alert>

        <n-card
            v-if="siteInfo && !loadError"
            title="运行概览"
            size="small"
            class="overview-card"
        >
            <template #header-extra>
                <n-tag
                    round
                    size="small"
                    :type="siteInfo.has_pending_restart ? 'warning' : 'success'"
                >
                    {{ siteInfo.has_pending_restart ? "存在待重启配置" : "无待重启配置" }}
                </n-tag>
            </template>

            <div class="overview-stats">
                <div class="overview-stat">
                    <span class="overview-stat-label">注册用户</span>
                    <strong class="overview-stat-value">
                        {{ siteInfo.register_user_count }}
                    </strong>
                </div>
                <div class="overview-stat">
                    <span class="overview-stat-label">当前在线</span>
                    <strong class="overview-stat-value">
                        {{ siteInfo.online_user_count }}
                    </strong>
                </div>
                <div class="overview-stat">
                    <span class="overview-stat-label">历史最高在线</span>
                    <strong class="overview-stat-value">
                        {{ siteInfo.history_max_online }}
                    </strong>
                </div>
                <div class="overview-stat">
                    <span class="overview-stat-label">当前适用配置项</span>
                    <strong class="overview-stat-value">
                        {{ siteInfo.active_settings_count }}
                    </strong>
                </div>
            </div>

            <div class="overview-details">
                <div class="overview-detail-item">
                    <span>本次启动时间</span>
                    <strong>{{ formatServerStart(siteInfo.server_up_time) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>版本</span>
                    <strong>{{ runtimeText(siteInfo.version) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>部署模式</span>
                    <strong>{{ runtimeText(siteInfo.deployment_mode) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>运行模式</span>
                    <strong>{{ runtimeText(siteInfo.run_mode) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>数据库</span>
                    <strong>{{ runtimeText(siteInfo.database_provider) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>缓存</span>
                    <strong>{{ runtimeText(siteInfo.cache_provider) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>搜索</span>
                    <strong>{{ runtimeText(siteInfo.search_provider) }}</strong>
                </div>
                <div class="overview-detail-item">
                    <span>存储</span>
                    <strong>{{ runtimeText(siteInfo.storage_provider) }}</strong>
                </div>
                <div class="overview-detail-item overview-detail-features">
                    <span>已启用功能</span>
                    <n-space
                        v-if="(siteInfo.enabled_features ?? []).length > 0"
                        size="small"
                    >
                        <n-tag
                            v-for="feature in siteInfo.enabled_features ?? []"
                            :key="feature"
                            round
                            size="small"
                        >
                            {{ feature }}
                        </n-tag>
                    </n-space>
                    <strong v-else>基础功能</strong>
                </div>
            </div>
        </n-card>

        <n-card v-if="!loadError" title="系统配置" size="small" class="setting-card">
            <n-spin :show="loading">
                <n-space vertical size="large" class="settings-layout">
                    <div class="settings-toolbar">
                        <n-input
                            v-model:value="searchKeyword"
                            clearable
                            round
                            placeholder="搜索中文名称、说明或配置 Key"
                        />
                        <span class="filter-result-text">
                            {{ filterResultText }}
                        </span>
                    </div>

                    <n-alert
                        v-if="hasPendingRestart"
                        type="warning"
                        class="setting-alert"
                    >
                        检测到部分配置已经保存，但需要重启服务后才会切换到新值。带有“待重启”标记的配置当前仍在使用旧的生效值。
                    </n-alert>

                    <div
                        v-if="!loading && filteredItemCount === 0"
                        class="empty-wrap"
                    >
                        <n-empty
                            size="large"
                            :description="
                                searchKeyword.trim()
                                    ? '没有匹配的配置项'
                                    : '暂无可显示的配置项'
                            "
                        />
                    </div>

                    <div
                        v-for="group in activeDomains"
                        :key="group.key"
                        class="domain-block"
                    >
                        <div class="domain-header">
                            <div>
                                <div class="domain-title">{{ group.label }}</div>
                                <div class="domain-subtitle">
                                    {{ group.itemCount }} 项配置
                                </div>
                            </div>
                            <n-tag round size="small">{{ group.key }}</n-tag>
                        </div>

                        <n-card
                            v-for="section in group.sections"
                            :key="`${group.key}-${section.key}`"
                            size="small"
                            class="section-card"
                            embedded
                            :bordered="false"
                        >
                            <template #header>
                                <div class="section-header">
                                    <div class="section-title">{{ section.label }}</div>
                                    <div class="section-subtitle">
                                        {{ section.items.length }} 项
                                    </div>
                                </div>
                            </template>

                            <div
                                v-for="entry in section.items"
                                :key="entry.schema.key"
                                class="setting-item"
                            >
                                <div class="setting-item-top">
                                    <div>
                                        <div class="setting-item-title">
                                            {{ displayLabel(entry.schema) }}
                                        </div>
                                        <div class="setting-item-key">
                                            {{ entry.schema.key }}
                                        </div>
                                    </div>
                                    <n-space size="small">
                                        <n-tag
                                            round
                                            size="small"
                                            :type="
                                                applyModeTagType(entry.schema.apply_mode)
                                            "
                                        >
                                            {{ applyModeLabel(entry.schema.apply_mode) }}
                                        </n-tag>
                                        <n-tag
                                            round
                                            size="small"
                                            :type="sourceTagType(entry.value?.source)"
                                        >
                                            {{ sourceLabel(entry.value?.source) }}
                                        </n-tag>
                                        <n-tag
                                            v-if="entry.schema.secret"
                                            round
                                            size="small"
                                        >
                                            机密
                                        </n-tag>
                                        <n-tag
                                            v-if="entry.value?.pending_restart"
                                            round
                                            size="small"
                                            type="warning"
                                        >
                                            待重启
                                        </n-tag>
                                        <n-tag
                                            v-if="
                                                entry.schema.readonly ||
                                                entry.schema.apply_mode ===
                                                    'bootstrap_only'
                                            "
                                            round
                                            size="small"
                                            type="default"
                                        >
                                            只读
                                        </n-tag>
                                    </n-space>
                                </div>

                                <div class="setting-item-description">
                                    {{ displayDescription(entry.schema) }}
                                </div>

                                <div class="setting-meta">
                                    <span class="meta-item">
                                        当前状态：{{ currentStatusText(entry) }}
                                    </span>
                                    <span class="meta-item">
                                        启动基线：{{ bootstrapStatusText(entry.schema) }}
                                    </span>
                                    <span
                                        v-if="showEffectiveValue(entry)"
                                        class="meta-item"
                                    >
                                        当前生效：{{
                                            formatValue(
                                                entry.value?.effective_value,
                                                entry.schema
                                            )
                                        }}
                                    </span>
                                </div>

                                <div class="setting-editor">
                                    <template v-if="entry.schema.secret">
                                        <n-input
                                            v-model:value="
                                                secretDraftValues[entry.schema.key]
                                            "
                                            type="password"
                                            show-password-on="click"
                                            :disabled="!isEditable(entry.schema)"
                                            placeholder="留空表示保持当前配置，输入后将替换"
                                        />
                                    </template>
                                    <template v-else-if="entry.schema.type === 'bool'">
                                        <n-switch
                                            :value="booleanDraftValue(entry.schema.key)"
                                            @update:value="
                                                (value) =>
                                                    setDraftValue(
                                                        entry.schema.key,
                                                        value
                                                    )
                                            "
                                            :disabled="!isEditable(entry.schema)"
                                        />
                                    </template>
                                    <template v-else-if="hasOptions(entry.schema)">
                                        <n-select
                                            :value="selectDraftValue(entry.schema.key)"
                                            @update:value="
                                                (value) =>
                                                    setDraftValue(
                                                        entry.schema.key,
                                                        value
                                                    )
                                            "
                                            :options="settingOptions(entry.schema)"
                                            :disabled="!isEditable(entry.schema)"
                                        />
                                    </template>
                                    <template
                                        v-else-if="
                                            entry.schema.type === 'int' ||
                                            entry.schema.type === 'float'
                                        "
                                    >
                                        <n-input-number
                                            :value="numericDraftValue(entry.schema.key)"
                                            @update:value="
                                                (value) =>
                                                    setDraftValue(
                                                        entry.schema.key,
                                                        value
                                                    )
                                            "
                                            class="number-input"
                                            :precision="
                                                entry.schema.type === 'float' ? 4 : 0
                                            "
                                            :step="
                                                entry.schema.type === 'float' ? 0.01 : 1
                                            "
                                            :disabled="!isEditable(entry.schema)"
                                        />
                                    </template>
                                    <template v-else>
                                        <n-input
                                            :value="stringDraftValue(entry.schema.key)"
                                            @update:value="
                                                (value) =>
                                                    setDraftValue(
                                                        entry.schema.key,
                                                        value
                                                    )
                                            "
                                            :type="
                                                useTextarea(entry.schema)
                                                    ? 'textarea'
                                                    : 'text'
                                            "
                                            :autosize="
                                                useTextarea(entry.schema)
                                                    ? { minRows: 3, maxRows: 6 }
                                                    : undefined
                                            "
                                            :disabled="!isEditable(entry.schema)"
                                            placeholder="请输入配置值"
                                        />
                                    </template>
                                </div>

                                <div class="setting-hint">
                                    {{ editorHintText(entry) }}
                                </div>
                            </div>
                        </n-card>
                    </div>

                    <n-collapse
                        v-if="inactiveDomains.length > 0"
                        class="inactive-collapse"
                    >
                        <n-collapse-item
                            :title="`未激活配置 (${inactiveItemCount})`"
                            name="inactive"
                        >
                            <div
                                v-for="group in inactiveDomains"
                                :key="`inactive-${group.key}`"
                                class="inactive-group"
                            >
                                <div class="inactive-group-title">{{ group.label }}</div>
                                <div
                                    v-for="section in group.sections"
                                    :key="`inactive-${group.key}-${section.key}`"
                                    class="inactive-section"
                                >
                                    <div class="inactive-section-title">
                                        {{ section.label }}
                                    </div>
                                    <div
                                        v-for="entry in section.items"
                                        :key="`inactive-${entry.schema.key}`"
                                        class="inactive-item"
                                    >
                                        <div class="inactive-item-top">
                                            <div>
                                                <div class="setting-item-title">
                                                    {{ displayLabel(entry.schema) }}
                                                </div>
                                                <div class="setting-item-key">
                                                    {{ entry.schema.key }}
                                                </div>
                                            </div>
                                            <n-space size="small">
                                                <n-tag round size="small" type="default"
                                                    >未激活</n-tag
                                                >
                                                <n-tag
                                                    round
                                                    size="small"
                                                    :type="
                                                        applyModeTagType(
                                                            entry.schema.apply_mode
                                                        )
                                                    "
                                                >
                                                    {{
                                                        applyModeLabel(
                                                            entry.schema.apply_mode
                                                        )
                                                    }}
                                                </n-tag>
                                                <n-tag
                                                    v-if="entry.schema.secret"
                                                    round
                                                    size="small"
                                                    >机密</n-tag
                                                >
                                            </n-space>
                                        </div>
                                        <div class="setting-item-description">
                                            {{ displayDescription(entry.schema) }}
                                        </div>
                                        <div class="setting-meta inactive-meta">
                                            <span class="meta-item">
                                                当前状态：{{ currentStatusText(entry) }}
                                            </span>
                                            <span class="meta-item">
                                                启动基线：{{
                                                    bootstrapStatusText(entry.schema)
                                                }}
                                            </span>
                                        </div>
                                        <div class="setting-hint">
                                            当前功能或能力未启用，所以该配置暂不参与当前运行时行为。
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </n-collapse-item>
                    </n-collapse>

                    <div class="form-submit-wrap">
                        <n-button
                            round
                            quaternary
                            :disabled="saving || loading || changedCount === 0"
                            @click="handleReset"
                        >
                            重置更改（{{ changedCount }}）
                        </n-button>
                        <n-button
                            round
                            type="primary"
                            secondary
                            :loading="saving"
                            :disabled="loading || changedCount === 0"
                            @click="handleSave"
                        >
                            保存配置（{{ changedCount }}）
                        </n-button>
                    </div>
                </n-space>
            </n-spin>
        </n-card>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { storeToRefs } from "pinia";
import { useRouter } from "vue-router";
import { userInfo as fetchUserInfo } from "@/api/auth";
import { getSiteProfile } from "@/api/site";
import { useStoreMain } from "@/store/main";
import { useStoreProfile } from "@/store/profile";
import { TOKEN_KEY, useStoreUser } from "@/store/user";
import { formatTime } from "@/utils/formatTime";
import { Api } from "@/utils/request";

type SettingPrimitive = string | number | boolean | null;
type SchemaItem = Api.Admin.NetReq.SettingsSchemaItem;
type ValueItem = Api.Admin.NetReq.SettingsValueItem;
type SiteInfo = Api.Admin.NetReq.SiteInfoResp;
type ViewItem = {
    schema: SchemaItem;
    value?: ValueItem;
};
type GroupedSection = {
    key: string;
    label: string;
    items: ViewItem[];
};
type GroupedDomain = {
    key: string;
    label: string;
    sections: GroupedSection[];
    itemCount: number;
};

const groupLabelMap: Record<string, string> = {
    web: "Web 站点",
    app: "应用行为",
    search: "搜索能力",
    storage: "对象存储",
    notifications: "通知服务",
    payments: "支付服务",
};

const sectionLabelMap: Record<string, string> = {
    profile: "站点资料",
    general: "常规设置",
    limits: "限制与阈值",
    bridge: "索引桥接",
    meili: "Meilisearch",
    zinc: "Zinc",
    common: "通用存储",
    local_oss: "本地 OSS",
    minio: "MinIO",
    s3: "Amazon S3",
    alioss: "AliOSS",
    cos: "腾讯 COS",
    huawei_obs: "华为 OBS",
    sms_juhe: "聚合短信",
    alipay: "支付宝",
};

const settingCopyMap: Record<string, [string, string]> = {
    "meili.api_key": [
        "Meilisearch 主密钥",
        "配置应用连接 Meilisearch 时使用的主密钥。",
    ],
    "web_profile.use_friendship": [
        "好友关系模式",
        "开启后，前台使用双向好友关系功能。",
    ],
    "web_profile.enable_trends_bar": [
        "显示趋势侧栏",
        "控制网页端是否展示趋势与热门内容侧栏。",
    ],
    "web_profile.enable_wallet": [
        "显示钱包入口",
        "控制前台钱包入口的显示，不等同于支付服务总开关。",
    ],
    "web_profile.allow_tweet_attachment": [
        "允许帖子附件",
        "控制前台是否允许在帖子中添加文件附件。",
    ],
    "web_profile.allow_tweet_attachment_price": [
        "允许付费附件",
        "控制前台是否允许为帖子附件设置价格。",
    ],
    "web_profile.allow_tweet_video": [
        "允许视频帖子",
        "控制前台是否显示帖子视频上传功能。",
    ],
    "web_profile.allow_user_register": [
        "允许用户注册",
        "关闭后，注册接口会立即停止接受新用户；已登录用户不受影响。",
    ],
    "web_profile.allow_phone_bind": [
        "允许绑定手机号",
        "关闭后，验证码发送和手机号绑定接口会立即停止使用。",
    ],
    "web_profile.default_tweet_max_length": [
        "帖子最大字数",
        "设置单条帖子允许输入的最大字数。",
    ],
    "web_profile.tweet_web_ellipsis_size": [
        "网页摘要字数",
        "设置网页信息流中帖子摘要的截断长度。",
    ],
    "web_profile.tweet_mobile_ellipsis_size": [
        "移动端摘要字数",
        "设置移动端信息流中帖子摘要的截断长度。",
    ],
    "web_profile.default_tweet_visibility": [
        "默认帖子可见范围",
        "设置用户新建帖子时默认选择的可见范围。",
    ],
    "web_profile.default_msg_loop_interval": [
        "消息轮询间隔",
        "设置前台检查未读消息的间隔，单位为毫秒。",
    ],
    "web_profile.copyright_top": [
        "版权主文字",
        "设置前台页脚顶部显示的版权文字。",
    ],
    "web_profile.copyright_left": [
        "左侧页脚文字",
        "设置前台页脚左侧链接的显示文字。",
    ],
    "web_profile.copyright_left_link": [
        "左侧页脚链接",
        "设置前台页脚左侧文字跳转的链接。",
    ],
    "web_profile.copyright_right": [
        "右侧页脚文字",
        "设置前台页脚右侧链接的显示文字。",
    ],
    "web_profile.copyright_right_link": [
        "右侧页脚链接",
        "设置前台页脚右侧文字跳转的链接。",
    ],
    "app.max_comment_count": [
        "单帖评论上限",
        "设置一条帖子最多允许保留的评论数量。",
    ],
    "app.attachment_income_rate": [
        "附件收入分成比例",
        "设置付费附件收入结算给作者的比例。",
    ],
    "app.default_page_size": [
        "默认分页数量",
        "设置列表接口每页默认返回的记录数。",
    ],
    "app.max_page_size": [
        "最大分页数量",
        "设置列表接口单页允许请求的最大记录数。",
    ],
    "app.max_whisper_daily": [
        "每日私信上限",
        "设置单个用户每天允许发送的私信数量。",
    ],
    "app.max_captcha_times": [
        "验证码请求上限",
        "设置验证码请求次数阈值。",
    ],
    "tweet_search.max_update_qps": [
        "搜索索引更新速率",
        "设置搜索索引后台更新的每秒处理上限。",
    ],
    "tweet_search.min_worker": [
        "搜索索引工作线程",
        "设置搜索索引后台任务的最小工作线程数。",
    ],
    "object_storage.retain_in_days": [
        "临时对象保留天数",
        "设置临时上传对象的保留时间。",
    ],
    "object_storage.temp_dir": [
        "临时对象目录",
        "设置对象存储使用的临时目录或前缀。",
    ],
};

const settingFieldLabelMap: Record<string, string> = {
    host: "服务地址",
    index: "索引名称",
    api_key: "API 密钥",
    secure: "启用 HTTPS",
    user: "用户名",
    password: "密码",
    save_path: "本地保存路径",
    bucket: "存储桶",
    domain: "访问域名",
    access_key: "Access Key",
    secret_key: "Secret Key",
    endpoint: "服务端点",
    access_key_id: "Access Key ID",
    access_key_secret: "Access Key Secret",
    secret_id: "Secret ID",
    region: "区域",
    gateway: "短信网关",
    key: "短信密钥",
    tpl_id: "短信模板 ID",
    tpl_val: "短信模板参数",
    app_id: "应用 ID",
    private_key: "应用私钥",
    root_cert_file: "支付宝根证书文件",
    public_cert_file: "支付宝公钥证书文件",
    app_public_cert_file: "应用公钥证书文件",
    in_production: "使用生产环境",
};

const visibilityLabelMap: Record<string, string> = {
    public: "公开",
    following: "关注可见",
    friend: "好友可见",
    private: "仅自己",
};

const storeMain = useStoreMain();
const storeUser = useStoreUser();
const storeProfile = useStoreProfile();
const { userInfo } = storeToRefs(storeUser);
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const loadError = ref("");
const hasPendingRestart = ref(false);
const searchKeyword = ref("");
const siteInfo = ref<SiteInfo | null>(null);
const schemaItems = ref<SchemaItem[]>([]);
const valueItems = ref<ValueItem[]>([]);
const draftValues = reactive<Record<string, SettingPrimitive>>({});
const initialDraftValues = reactive<Record<string, SettingPrimitive>>({});
const secretDraftValues = reactive<Record<string, string>>({});

const valueMap = computed(() => {
    const map: Record<string, ValueItem> = {};
    for (const item of valueItems.value) {
        map[item.key] = item;
    }
    return map;
});

const settingFieldKey = (schema: SchemaItem) => {
    const separatorIndex = schema.key.indexOf(".");
    return separatorIndex >= 0 ? schema.key.slice(separatorIndex + 1) : schema.key;
};

const displayLabel = (schema: SchemaItem) => {
    const localized = settingCopyMap[schema.key];
    if (localized) {
        return localized[0];
    }

    const fieldLabel = settingFieldLabelMap[settingFieldKey(schema)];
    if (fieldLabel) {
        const section = sectionLabelMap[schema.section] ?? schema.section;
        return `${section} · ${fieldLabel}`;
    }
    return schema.label;
};

const displayDescription = (schema: SchemaItem) => {
    if (schema.key === "web_profile.allow_phone_bind" && !schema.active) {
        return "当前部署未启用 Sms 功能，后端会拒绝发送验证码和绑定手机号；启用并配置可靠的短信服务后才能开放。";
    }

    if (schema.key === "meili.api_key") {
        return schema.readonly || schema.apply_mode === "bootstrap_only"
            ? "当前由 .env 的 PAOPAO_MEILI_MASTER_KEY 托管；修改后重新执行 manage.sh up，不能只在后台单边更换。"
            : "保存新的 Meilisearch 主密钥后，需要重启应用才能生效，并确保搜索服务使用同一密钥。";
    }

    if (
        schema.key === "local_oss.domain" ||
        schema.key === "local_oss.secure"
    ) {
        return schema.readonly || schema.apply_mode === "bootstrap_only"
            ? "托管部署由 .env 的 PAOPAO_PUBLIC_HOST / PAOPAO_PUBLIC_SECURE 统一控制；修改后重新执行对应模式的 manage.sh up。"
            : schema.description;
    }

    const localized = settingCopyMap[schema.key];
    if (localized) {
        return localized[1];
    }

    const fieldLabel = settingFieldLabelMap[settingFieldKey(schema)];
    if (fieldLabel) {
        const section = sectionLabelMap[schema.section] ?? schema.section;
        return `配置 ${section} 使用的${fieldLabel}。`;
    }
    return schema.description;
};

const normalizedSearchKeyword = computed(() =>
    searchKeyword.value.trim().toLocaleLowerCase()
);

const matchesSearch = (schema: SchemaItem) => {
    const keyword = normalizedSearchKeyword.value;
    if (!keyword) {
        return true;
    }

    const searchableText = [
        displayLabel(schema),
        displayDescription(schema),
        schema.label,
        schema.description,
        schema.key,
        groupLabelMap[schema.group],
        sectionLabelMap[schema.section],
    ]
        .filter(Boolean)
        .join(" ")
        .toLocaleLowerCase();
    return searchableText.includes(keyword);
};

const activeDomains = computed(() => buildDomains(true));
const inactiveDomains = computed(() => buildDomains(false));
const inactiveItemCount = computed(() => {
    return inactiveDomains.value.reduce((count, domain) => count + domain.itemCount, 0);
});
const filteredItemCount = computed(() => {
    const activeCount = activeDomains.value.reduce(
        (count, domain) => count + domain.itemCount,
        0
    );
    return activeCount + inactiveItemCount.value;
});
const filterResultText = computed(() => {
    if (normalizedSearchKeyword.value) {
        return `找到 ${filteredItemCount.value} 项（共 ${schemaItems.value.length} 项）`;
    }
    return `共 ${schemaItems.value.length} 项配置`;
});

const buildDomains = (active: boolean): GroupedDomain[] => {
    const domains: GroupedDomain[] = [];
    const domainMap: Record<string, GroupedDomain> = {};

    for (const schema of schemaItems.value) {
        if (schema.active !== active || !matchesSearch(schema)) {
            continue;
        }

        if (!domainMap[schema.group]) {
            domainMap[schema.group] = {
                key: schema.group,
                label: groupLabel(schema.group),
                sections: [],
                itemCount: 0,
            };
            domains.push(domainMap[schema.group]);
        }

        const domain = domainMap[schema.group];
        let section = domain.sections.find((item) => item.key === schema.section);
        if (!section) {
            section = {
                key: schema.section,
                label: sectionLabel(schema.section),
                items: [],
            };
            domain.sections.push(section);
        }

        section.items.push({
            schema,
            value: valueMap.value[schema.key],
        });
        domain.itemCount++;
    }

    return domains;
};

const replaceRecord = <T>(target: Record<string, T>, next: Record<string, T>) => {
    for (const key of Object.keys(target)) {
        delete target[key];
    }
    for (const [key, value] of Object.entries(next)) {
        target[key] = value;
    }
};

const prettifyKey = (value: string) => {
    return value
        .split("_")
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
        .join(" ");
};

const groupLabel = (group: string) => {
    return groupLabelMap[group] ?? prettifyKey(group);
};

const sectionLabel = (section: string) => {
    return sectionLabelMap[section] ?? prettifyKey(section);
};

const runtimeText = (value?: string) => {
    const text = value?.trim();
    return text ? text : "未知";
};

const formatServerStart = (timestamp: number) => {
    return timestamp > 0 ? formatTime(timestamp) : "未知";
};

const isEditable = (schema: SchemaItem) => {
    return schema.active && !schema.readonly && schema.apply_mode !== "bootstrap_only";
};

const hasOptions = (schema: SchemaItem) => {
    return !!schema.options && schema.options.length > 0;
};

const settingOptions = (
    schema: SchemaItem
): Array<{ label: string; value: string | number }> => {
    return (schema.options ?? []).flatMap((option) => {
        if (typeof option.value !== "string" && typeof option.value !== "number") {
            return [];
        }

        return [
            {
                label:
                    schema.key === "web_profile.default_tweet_visibility"
                        ? visibilityLabelMap[String(option.value)] ?? option.label
                        : option.label,
                value: option.value,
            },
        ];
    });
};

const sourceLabel = (source?: string) => {
    return source === "override" ? "管理覆盖" : "启动配置";
};

const sourceTagType = (source?: string) => {
    return source === "override" ? "success" : "default";
};

const applyModeLabel = (applyMode: SchemaItem["apply_mode"]) => {
    switch (applyMode) {
        case "live":
            return "即时生效";
        case "restart_required":
            return "重启后生效";
        case "bootstrap_only":
            return "仅启动配置";
        default:
            return applyMode;
    }
};

const applyModeTagType = (applyMode: SchemaItem["apply_mode"]) => {
    switch (applyMode) {
        case "live":
            return "success";
        case "restart_required":
            return "warning";
        case "bootstrap_only":
            return "default";
        default:
            return "default";
    }
};

const fallbackValueForType = (schema: SchemaItem): SettingPrimitive => {
    switch (schema.type) {
        case "bool":
            return false;
        case "int":
        case "float":
            return null;
        default:
            return "";
    }
};

const normalizeDraftValue = (
    schema: SchemaItem,
    value: SettingPrimitive
): SettingPrimitive => {
    if (value === undefined) {
        return fallbackValueForType(schema);
    }

    switch (schema.type) {
        case "bool":
            return Boolean(value);
        case "int":
        case "float":
            if (value === null || value === "") {
                return null;
            }
            return typeof value === "number" ? value : Number(value);
        default:
            return typeof value === "string" ? value : String(value ?? "");
    }
};

const isDraftChanged = (schema: SchemaItem) => {
    if (!isEditable(schema)) {
        return false;
    }
    if (schema.secret) {
        return (secretDraftValues[schema.key] ?? "").trim().length > 0;
    }

    const nextValue = normalizeDraftValue(schema, draftValues[schema.key]);
    const initialValue = normalizeDraftValue(schema, initialDraftValues[schema.key]);
    if (schema.type === "string") {
        return String(nextValue ?? "").trim() !== String(initialValue ?? "").trim();
    }
    return nextValue !== initialValue;
};

const changedCount = computed(
    () => schemaItems.value.filter((schema) => isDraftChanged(schema)).length
);

const extractDraftValue = (schema: SchemaItem): SettingPrimitive => {
    const current = valueMap.value[schema.key];
    if (schema.secret) {
        return fallbackValueForType(schema);
    }
    if (current?.value !== undefined) {
        return normalizeDraftValue(schema, current.value as SettingPrimitive);
    }
    if (current?.effective_value !== undefined) {
        return normalizeDraftValue(schema, current.effective_value as SettingPrimitive);
    }
    if (schema.bootstrap_value !== undefined) {
        return normalizeDraftValue(schema, schema.bootstrap_value as SettingPrimitive);
    }
    return fallbackValueForType(schema);
};

const rebuildDraftState = () => {
    const nextDrafts: Record<string, SettingPrimitive> = {};
    const nextInitialDrafts: Record<string, SettingPrimitive> = {};
    const nextSecretDrafts: Record<string, string> = {};

    for (const schema of schemaItems.value) {
        if (schema.secret) {
            nextSecretDrafts[schema.key] = "";
            continue;
        }

        const draftValue = extractDraftValue(schema);
        nextDrafts[schema.key] = draftValue;
        nextInitialDrafts[schema.key] = draftValue;
    }

    replaceRecord(draftValues, nextDrafts);
    replaceRecord(initialDraftValues, nextInitialDrafts);
    replaceRecord(secretDraftValues, nextSecretDrafts);
};

const formatValue = (value: unknown, schema?: SchemaItem) => {
    if (value === null || value === undefined) {
        return "未设置";
    }
    if (typeof value === "boolean") {
        return value ? "已开启" : "已关闭";
    }
    if (typeof value === "number") {
        if (schema?.type === "float") {
            return `${value}`;
        }
        return `${value}`;
    }
    const text = String(value).trim();
    if (schema?.key === "web_profile.default_tweet_visibility") {
        return visibilityLabelMap[text] ?? text;
    }
    return text.length > 0 ? text : "未设置";
};

const configuredText = (configured?: boolean) => {
    return configured ? "已配置" : "未配置";
};

const currentConfiguredValue = (entry: ViewItem) => {
    if (entry.value?.value !== undefined) {
        return entry.value.value;
    }
    if (entry.value?.effective_value !== undefined) {
        return entry.value.effective_value;
    }
    if (entry.schema.bootstrap_value !== undefined) {
        return entry.schema.bootstrap_value;
    }
    return null;
};

const currentStatusText = (entry: ViewItem) => {
    if (entry.schema.secret) {
        return configuredText(
            entry.value?.configured ?? entry.schema.bootstrap_configured
        );
    }
    return formatValue(currentConfiguredValue(entry), entry.schema);
};

const bootstrapStatusText = (schema: SchemaItem) => {
    if (schema.secret) {
        return configuredText(schema.bootstrap_configured);
    }
    return formatValue(schema.bootstrap_value, schema);
};

const showEffectiveValue = (entry: ViewItem) => {
    return (
        !entry.schema.secret &&
        entry.value?.pending_restart &&
        entry.value.effective_value !== undefined &&
        entry.value.value !== undefined &&
        entry.value.effective_value !== entry.value.value
    );
};

const useTextarea = (schema: SchemaItem) => {
    return schema.key.includes("private_key");
};

const editorHintText = (entry: ViewItem) => {
    if (!entry.schema.active) {
        return "当前功能未启用，所以这项配置暂不参与当前运行行为。";
    }
    if (entry.schema.secret) {
        if (!isEditable(entry.schema)) {
            return "为安全起见不会显示当前明文，当前项仅展示是否已配置。";
        }
        return "为安全起见不会回显当前明文。留空表示保持原值，输入新内容后会安全替换。";
    }
    if (entry.schema.apply_mode === "bootstrap_only" || entry.schema.readonly) {
        return "该项由启动配置控制，当前页面仅提供查看，不允许直接覆盖。";
    }
    if (entry.value?.pending_restart) {
        return "新配置已保存，但当前服务仍在使用旧值；重启后会切换到已保存的配置。";
    }
    if (entry.schema.apply_mode === "restart_required") {
        return "保存后会写入管理覆盖值，但需要服务重启后才会切换。";
    }
    return "保存后会立即刷新当前服务中的配置状态。";
};

const refreshPublicSiteProfile = async (updatedKeys: string[]) => {
    if (!updatedKeys.some((key) => key.startsWith("web_profile."))) {
        return;
    }

    try {
        const profile = await getSiteProfile();
        storeProfile.updateSiteProfile(profile);
    } catch (_err) {
        // The request layer displays the backend error. Keep the current draft
        // intact so the administrator can correct it and retry.
    }
};

const readableLoadError = (error: unknown) => {
    if (typeof error === "string" && error.trim()) {
        return error.trim();
    }
    if (error && typeof error === "object") {
        const detail = error as Record<string, unknown>;
        if (
            Array.isArray(detail.details) &&
            typeof detail.details[0] === "string" &&
            detail.details[0].trim()
        ) {
            return detail.details[0].trim();
        }
        if (typeof detail.msg === "string" && detail.msg.trim()) {
            return detail.msg.trim();
        }
        if (typeof detail.message === "string" && detail.message.trim()) {
            return detail.message.trim();
        }
    }
    return "暂时无法读取运行状态或系统配置，请检查服务连接后重试。";
};

const loadSettings = async () => {
    loading.value = true;
    loadError.value = "";
    try {
        const [siteResp, schemaResp, valuesResp] = await Promise.all([
            Api.v1.admin.get.site.status(),
            Api.v1.admin.get.settings.schema(),
            Api.v1.admin.get.settings.values(),
        ]);
        siteInfo.value = siteResp;
        schemaItems.value = schemaResp.items;
        valueItems.value = valuesResp.items;
        hasPendingRestart.value =
            valuesResp.has_pending_restart || siteResp.has_pending_restart;
        rebuildDraftState();
    } catch (error) {
        loadError.value = readableLoadError(error);
    } finally {
        loading.value = false;
    }
};

const ensureAdminAccess = async () => {
    if (!localStorage.getItem(TOKEN_KEY) && userInfo.value.id === 0) {
        storeMain.triggerAuth(true);
        storeMain.triggerAuthKey("signin");
        router.replace({
            name: "home",
        });
        return false;
    }

    if (userInfo.value.id === 0) {
        try {
            const currentUser = await fetchUserInfo();
            storeUser.updateUserinfo(currentUser);
        } catch (_err) {
            storeUser.userLogout();
            router.replace({
                name: "home",
            });
            return false;
        }
    }

    if (!userInfo.value.is_admin) {
        router.replace({
            name: "404",
        });
        return false;
    }

    return true;
};

const collectChangedItems = (): Api.Admin.NetParams.SettingValueInput[] | null => {
    normalizeDependentDrafts();

    const items: Api.Admin.NetParams.SettingValueInput[] = [];

    for (const schema of schemaItems.value) {
        if (!isEditable(schema)) {
            continue;
        }

        if (schema.secret) {
            const secretValue = (secretDraftValues[schema.key] ?? "").trim();
            if (secretValue.length > 0) {
                items.push({
                    key: schema.key,
                    value: secretValue,
                });
            }
            continue;
        }

        const nextValue = normalizeDraftValue(schema, draftValues[schema.key]);
        const initialValue = normalizeDraftValue(schema, initialDraftValues[schema.key]);

        if (
            (schema.type === "int" || schema.type === "float") &&
            (nextValue === null || Number.isNaN(nextValue))
        ) {
            window.$message.warning(`${displayLabel(schema)} 请输入有效数值`);
            return null;
        }

        if (schema.type === "string") {
            const nextText = String(nextValue ?? "").trim();
            const initialText = String(initialValue ?? "").trim();
            if (nextText !== initialText) {
                items.push({
                    key: schema.key,
                    value: nextText,
                });
            }
            continue;
        }

        if (nextValue !== initialValue) {
            items.push({
                key: schema.key,
                value: nextValue as string | number | boolean,
            });
        }
    }

    return items;
};

const numericDraftValue = (key: string): number | null => {
    const schema = schemaItems.value.find((item) => item.key === key);
    if (!schema) {
        return null;
    }
    const value = normalizeDraftValue(schema, draftValues[key]);
    if (typeof value !== "number" || Number.isNaN(value)) {
        return null;
    }
    return value;
};

const booleanDraftValue = (key: string): boolean => {
    return draftValues[key] === true;
};

const selectDraftValue = (key: string): string | number | null => {
    const value = draftValues[key];
    return typeof value === "string" || typeof value === "number" ? value : null;
};

const stringDraftValue = (key: string): string => {
    const value = draftValues[key];
    return value === null || value === undefined ? "" : String(value);
};

const setDraftValue = (key: string, value: SettingPrimitive | undefined) => {
    draftValues[key] = value ?? null;
};

const setNumericDraftIfPresent = (key: string, nextValue: number) => {
    if (!(key in draftValues)) {
        return;
    }
    draftValues[key] = nextValue;
};

const normalizeDependentDrafts = () => {
    const maxTweetLength = numericDraftValue("web_profile.default_tweet_max_length");
    if (maxTweetLength !== null) {
        const webEllipsis = numericDraftValue("web_profile.tweet_web_ellipsis_size");
        const mobileEllipsis = numericDraftValue("web_profile.tweet_mobile_ellipsis_size");
        if (webEllipsis !== null && webEllipsis > maxTweetLength) {
            setNumericDraftIfPresent("web_profile.tweet_web_ellipsis_size", maxTweetLength);
        }
        if (mobileEllipsis !== null && mobileEllipsis > maxTweetLength) {
            setNumericDraftIfPresent("web_profile.tweet_mobile_ellipsis_size", maxTweetLength);
        }
    }

    const maxPageSize = numericDraftValue("app.max_page_size");
    const defaultPageSize = numericDraftValue("app.default_page_size");
    if (maxPageSize !== null && defaultPageSize !== null && defaultPageSize > maxPageSize) {
        setNumericDraftIfPresent("app.default_page_size", maxPageSize);
    }
};

const handleReset = () => {
    rebuildDraftState();
};

const handleSave = async (e: MouseEvent) => {
    e.preventDefault();

    const items = collectChangedItems();
    if (!items) {
        return;
    }
    if (items.length === 0) {
        window.$message.info("没有需要保存的变更");
        return;
    }

    saving.value = true;
    try {
        const resp = await Api.v1.admin.post.settings.save({ items });
        valueItems.value = resp.items;
        hasPendingRestart.value = resp.has_pending_restart;
        if (siteInfo.value) {
            siteInfo.value.has_pending_restart = resp.has_pending_restart;
        }
        rebuildDraftState();
        await refreshPublicSiteProfile(resp.updated_keys);
        window.$message.success(`已保存 ${resp.updated_keys.length} 项配置`);
    } catch (_err) {
        // do nothing
    } finally {
        saving.value = false;
    }
};

onMounted(async () => {
    const allowed = await ensureAdminAccess();
    if (!allowed) {
        return;
    }
    await loadSettings();
});
</script>

<style lang="less" scoped>
.load-error-alert,
.overview-card {
    margin-top: -1px;
    border-radius: 0;
}

.load-error-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.overview-card {
    .overview-stats {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 10px;
    }

    .overview-stat {
        display: flex;
        flex-direction: column;
        gap: 5px;
        min-width: 0;
        padding: 12px;
        border-radius: 10px;
        background-color: rgba(127, 127, 127, 0.08);

        .overview-stat-label {
            font-size: 12px;
            opacity: 0.68;
        }

        .overview-stat-value {
            overflow: hidden;
            font-size: 22px;
            line-height: 1.2;
            text-overflow: ellipsis;
        }
    }

    .overview-details {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 10px 24px;
        margin-top: 16px;
    }

    .overview-detail-item {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 12px;
        min-width: 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(127, 127, 127, 0.12);

        span {
            flex: 0 0 auto;
            font-size: 12px;
            opacity: 0.65;
        }

        strong {
            overflow-wrap: anywhere;
            text-align: right;
        }
    }

    .overview-detail-features {
        grid-column: 1 / -1;
    }
}

.setting-card {
    margin-top: 10px;
    border-radius: 0;

    .settings-layout {
        width: 100%;
    }

    .setting-alert {
        margin-bottom: 0;
    }

    .settings-toolbar {
        display: flex;
        align-items: center;
        gap: 12px;

        .filter-result-text {
            flex: 0 0 auto;
            font-size: 12px;
            opacity: 0.68;
        }
    }

    .domain-block {
        display: flex;
        flex-direction: column;
        gap: 12px;

        .domain-header {
            display: flex;
            justify-content: space-between;
            align-items: center;

            .domain-title {
                font-size: 16px;
                font-weight: 600;
            }

            .domain-subtitle {
                font-size: 12px;
                opacity: 0.7;
                margin-top: 4px;
            }
        }
    }

    .section-card {
        border-radius: 12px;

        .section-header {
            display: flex;
            justify-content: space-between;
            align-items: baseline;

            .section-title {
                font-weight: 600;
            }

            .section-subtitle {
                font-size: 12px;
                opacity: 0.7;
            }
        }
    }

    .setting-item,
    .inactive-item {
        padding: 8px 0 18px;
        border-bottom: 1px solid rgba(127, 127, 127, 0.12);

        &:last-child {
            padding-bottom: 0;
            border-bottom: none;
        }

        .setting-item-top,
        .inactive-item-top {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            gap: 12px;
        }

        .setting-item-title {
            font-size: 15px;
            font-weight: 600;
        }

        .setting-item-key {
            margin-top: 4px;
            font-size: 12px;
            opacity: 0.55;
            word-break: break-all;
        }

        .setting-item-description {
            margin-top: 8px;
            font-size: 13px;
            line-height: 1.7;
            opacity: 0.82;
        }

        .setting-meta {
            display: flex;
            flex-wrap: wrap;
            gap: 8px 16px;
            margin-top: 10px;

            .meta-item {
                font-size: 12px;
                opacity: 0.72;
            }
        }

        .setting-editor {
            margin-top: 12px;

            .number-input {
                width: 220px;
            }
        }

        .setting-hint {
            margin-top: 8px;
            font-size: 12px;
            line-height: 1.7;
            opacity: 0.72;
        }
    }

    .inactive-collapse {
        margin-top: 4px;

        .inactive-group {
            display: flex;
            flex-direction: column;
            gap: 16px;
        }

        .inactive-group-title {
            font-size: 15px;
            font-weight: 600;
        }

        .inactive-section {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .inactive-section-title {
            font-size: 13px;
            font-weight: 600;
            opacity: 0.75;
        }

        .inactive-meta {
            margin-top: 8px;
        }
    }

    .empty-wrap {
        padding: 40px 0;
    }

    .form-submit-wrap {
        display: flex;
        justify-content: flex-end;
        gap: 12px;
        padding-top: 4px;
    }
}

.dark {
    .setting-card,
    .overview-card {
        background-color: rgba(16, 16, 20, 0.75);
    }

    .section-card {
        background-color: #18181c;
    }
}

@media screen and (max-width: 600px) {
    .load-error-content {
        align-items: flex-start;
        flex-direction: column;
    }

    .overview-card {
        .overview-stats {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }

        .overview-details {
            grid-template-columns: 1fr;
        }

        .overview-detail-features {
            grid-column: auto;
        }
    }

    .setting-card {
        .settings-toolbar {
            align-items: stretch;
            flex-direction: column;
            gap: 6px;
        }

        .setting-item,
        .inactive-item {
            .setting-item-top,
            .inactive-item-top {
                flex-direction: column;
            }

            .setting-editor .number-input {
                width: 100%;
            }
        }

        .form-submit-wrap {
            flex-wrap: wrap;
        }
    }
}
</style>
