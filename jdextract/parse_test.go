package jdextract

// sampleJD is a real jina.ai markdown response for a VML Copywriter posting fetched from LinkedIn.
// Source: https://www.vml.com/careers/job/8234798002-ca-copywriter?gh_jid=8234798002
// Used to surface real-world regex gaps rather than sanitised synthetic input.
const sampleJD = `Title: Copywriter | Careers | VML

URL Source: https://www.vml.com/careers/job/8234798002-ca-copywriter?gh_jid=8234798002

Markdown Content:
Copywriter
----------

#### **Brand:** VML

#### **Capability:** Creative

#### **Location:**Toronto, Canada

#### **Last Updated:**2/25/2026

#### **Requisition ID:**12108

### **About VML**

VML is a leading creative company that combines brand experience, customer experience,
and commerce, creating connected brands to drive growth.

We're seeking an Intermediate Copywriter to craft compelling, on-brand copy that drives
engagement and performance across integrated campaigns, product experiences, and content.

**Key Responsibilities**

*   Translate briefs into messaging strategies, narratives, and copy across digital,
    social, email, web, print, OOH, and video/radio.

*   Write headlines, body copy, taglines, CTAs, scripts, and website copy that reflect
    brand voice and meet business objectives.

*   Partner with art directors to concept integrated ideas; present work with clear
    rationale and incorporate feedback thoughtfully.

*   Edit and proof for grammar, style, and consistency; manage file/version control.

*   Manage multiple workstreams, estimate level of effort, and deliver on time and within scope.

**Qualifications**

*   3-4 years of professional copywriting experience (agency or in-house).

*   Portfolio showcasing concept-driven campaigns, digital-first copy, performance
    creative, and cohesive brand storytelling.

$65,000—$115,000 CAD

We believe the best work happens when we're together. That's why we've adopted a hybrid
approach, with teams in the office an average of four days a week.`

// sampleJD2 is a real jina.ai markdown response for a Felix Senior Copywriter posting on Ashby.
// Source: https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b
// Key structural difference from sampleJD: Ashby uses **bold paragraphs** for section headers,
// not markdown ATX headings. headingRe is completely blind to this document's structure.
const sampleJD2 = `Title: Senior Copywriter

URL Source: https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b

Markdown Content:
[Overview](https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b)[Application](https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b/application)

**About Felix**

Felix is Canada's first end-to-end platform providing on-demand treatment for everyday health.
Felix has delivered 24 quarters of consecutive growth in both revenue and profit.

**The Role**

We are seeking an experienced senior copywriter with a deep understanding of writing, marketing,
and creativity to join the Felix brand team.

**In this role, you will:**

*   Establish tone of voice, consistent terminology, and tonal considerations for the brand

*   Collaborate with the growth team to develop copy and messaging for static and video digital
    ad creative for always-on paid social marketing

*   Write scripts for both brand & functional level TVCs for mass-reach campaigns

*   Work closely with compliance teams to balance creativity with accuracy and compliance

**We're looking for someone who:**

*   Has 7+ years of writing experience, including 3+ years of in-house experience

*   Is comfortable co-writing and receiving creative briefs

*   Has a portfolio of relevant work experience

**Benefits**

*   Full medical, dental and vision benefits

*   Remote first, work from anywhere in Canada

*   Stock option grant

**Location:**Toronto, Remote (Canada). We have an office in Toronto, but are currently working
remotely and open to candidates from anywhere in Canada.`
